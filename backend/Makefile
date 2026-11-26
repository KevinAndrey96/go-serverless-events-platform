GO_BUILD = GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -trimpath -ldflags="-s -w"

.PHONY: test test-integration lint

test:
	go test -race -count=1 ./...

test-integration:
	DYNAMODB_ENDPOINT=$${DYNAMODB_ENDPOINT:-http://localhost:8000} go test -race -count=1 ./internal/dynamo/...

lint:
	golangci-lint run ./...

# sam build calls build-<LogicalId> and expects the binary in $(ARTIFACTS_DIR)/bootstrap.
build-ListEventsFunction:
	$(GO_BUILD) -o $(ARTIFACTS_DIR)/bootstrap ./cmd/list-events

build-RegisterAttendeeFunction:
	$(GO_BUILD) -o $(ARTIFACTS_DIR)/bootstrap ./cmd/register-attendee

build-ListEventsAdminFunction:
	$(GO_BUILD) -o $(ARTIFACTS_DIR)/bootstrap ./cmd/list-events-admin

build-CreateEventFunction:
	$(GO_BUILD) -o $(ARTIFACTS_DIR)/bootstrap ./cmd/create-event

build-ChangeEventStatusFunction:
	$(GO_BUILD) -o $(ARTIFACTS_DIR)/bootstrap ./cmd/change-event-status

build-ListAttendeesFunction:
	$(GO_BUILD) -o $(ARTIFACTS_DIR)/bootstrap ./cmd/list-attendees

build-ChangePaymentStatusFunction:
	$(GO_BUILD) -o $(ARTIFACTS_DIR)/bootstrap ./cmd/change-payment-status

build-CheckInFunction:
	$(GO_BUILD) -o $(ARTIFACTS_DIR)/bootstrap ./cmd/check-in

build-SendRemindersFunction:
	$(GO_BUILD) -o $(ARTIFACTS_DIR)/bootstrap ./cmd/send-reminders
