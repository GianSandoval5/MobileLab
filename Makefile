GO ?= go
BINARY := bin/mobilelab
VERSION ?= 0.1.0-dev
LDFLAGS ?= -s -w -X github.com/mobilelab-dev/mobilelab/internal/cli.Version=$(VERSION)

.PHONY: build test lint run clean release-check

build:
	@mkdir -p bin
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/mobilelab

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

run:
	$(GO) run ./cmd/mobilelab

release-check: test lint build
	@test "$$($(BINARY) --version)" = "MobileLab $(VERSION)"

clean:
	rm -rf bin coverage.out
