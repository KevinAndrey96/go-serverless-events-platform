# Local dev server: the Lambda handlers over plain HTTP, for Docker Compose.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/devserver ./cmd/devserver

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/devserver /devserver
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/devserver"]
