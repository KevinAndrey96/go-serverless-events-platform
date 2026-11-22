.PHONY: all test build clean lint

FUNCTIONS := create show change_status change_pay_status event-users reminder

all: test build

test:
	go test -v -race ./pkg/...

build: $(FUNCTIONS)

$(FUNCTIONS):
	@echo "Building $@..."
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -tags lambda.norpc -o bin/$@ $@/v1/main.go

clean:
	rm -rf bin/ bootstrap
