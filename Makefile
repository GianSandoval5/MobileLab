GO ?= go
BINARY := bin/mobilelab

.PHONY: build test lint run clean

build:
	@mkdir -p bin
	$(GO) build -trimpath -o $(BINARY) ./cmd/mobilelab

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

run:
	$(GO) run ./cmd/mobilelab

clean:
	rm -rf bin coverage.out
