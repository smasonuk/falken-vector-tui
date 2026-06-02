.PHONY: build test lint clean

GO ?= go
BINARY ?= bin/falken-vector-tui

build:
	mkdir -p $(dir $(BINARY))
	CGO_ENABLED=0 $(GO) build -o $(BINARY) ./cmd/falken-vector-tui

test:
	CGO_ENABLED=0 $(GO) test ./...

lint:
	$(GO) vet ./...

clean:
	rm -f $(BINARY)
