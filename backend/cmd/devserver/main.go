// Command devserver serves the same handlers as the Lambdas over plain HTTP, against DynamoDB Local.
//
// It has no authentication: it exists for local development and the Docker Compose stack only.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/dynamo"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/handler"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/mail"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/platform"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/usecase"
)

const (
	defaultAddr     = ":8080"
	maxRequestBytes = 64 << 10
	shutdownTimeout = 10 * time.Second
)

type lambdaHandler func(ctx context.Context, req handler.Request) (handler.Response, error)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		platform.Exit(err)
	}
}

func run(ctx context.Context) error {
	deps, err := platform.Bootstrap(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	if err := dynamo.EnsureTable(ctx, deps.Client, deps.Config.Table); err != nil {
		return fmt.Errorf("local table: %w", err)
	}
	addr := os.Getenv("DEV_ADDR")
	if addr == "" {
		addr = defaultAddr
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           routes(deps),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
	}
	served := make(chan error, 1)
	go func() { served <- server.ListenAndServe() }()
	deps.Logger.InfoContext(ctx, "devserver listening", "addr", addr, "table", deps.Config.Table)

	select {
	case err := <-served:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}

func routes(deps platform.Deps) *http.ServeMux {
	store, logger := deps.Store, deps.Logger
	listEvents := handler.NewListEvents(usecase.NewListEvents(store), logger)
	reminders := handler.NewSendReminders(usecase.NewSendReminders(store, mail.NewLog(logger), logger), logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.Handle("GET /events", adapt(listEvents.HandlePublic, logger))
	mux.Handle("POST /events/{eventId}/attendees",
		adapt(handler.NewRegisterAttendee(usecase.NewRegisterAttendee(store), logger).Handle, logger))
	mux.Handle("GET /admin/events", adapt(listEvents.HandleAdmin, logger))
	mux.Handle("POST /admin/events", adapt(handler.NewCreateEvent(usecase.NewCreateEvent(store), logger).Handle, logger))
	mux.Handle("PATCH /admin/events/{eventId}/status",
		adapt(handler.NewChangeEventStatus(usecase.NewChangeEventStatus(store), logger).Handle, logger))
	mux.Handle("GET /admin/events/{eventId}/attendees",
		adapt(handler.NewListAttendees(usecase.NewListAttendees(store), logger).Handle, logger))
	mux.Handle("PATCH /admin/events/{eventId}/attendees/{attendeeId}/payment",
		adapt(handler.NewChangePaymentStatus(usecase.NewChangePaymentStatus(store), logger).Handle, logger))
	mux.Handle("POST /admin/check-ins", adapt(handler.NewCheckIn(usecase.NewCheckIn(store), logger).Handle, logger))
	mux.HandleFunc("POST /admin/reminders", func(w http.ResponseWriter, r *http.Request) {
		report, err := reminders.Handle(r.Context())
		if err != nil {
			http.Error(w, "reminders failed", http.StatusInternalServerError)

			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"events":%d,"sent":%d,"skipped":%d,"failed":%d}`,
			report.Events, report.Sent, report.Skipped, report.Failed)
	})

	return mux
}

func adapt(fn lambdaHandler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBytes))
		if err != nil {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)

			return
		}
		params := map[string]string{}
		for _, name := range []string{"eventId", "attendeeId"} {
			if value := r.PathValue(name); value != "" {
				params[name] = value
			}
		}
		resp, err := fn(r.Context(), handler.Request{RawPath: r.URL.Path, Body: string(body), PathParameters: params})
		if err != nil {
			logger.ErrorContext(r.Context(), "handler", "path", r.URL.Path, "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)

			return
		}
		for name, value := range resp.Headers {
			w.Header().Set(name, value)
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.WriteString(w, resp.Body)
	})
}
