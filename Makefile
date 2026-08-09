GO ?= go
BINARY := bin/mobilelab
PLUGIN_EXAMPLE_BINARY := examples/plugins/mobilelab/plugins/echo/mobilelab-plugin-echo$(if $(filter Windows_NT,$(OS)),.exe,)
VERSION ?= 0.7.0-dev
LDFLAGS ?= -s -w -X github.com/mobilelab-dev/mobilelab/internal/cli.Version=$(VERSION)

.PHONY: build test lint run clean release-check ci-example plugin-example plugin-example-check sdk-test sdk-flutter-test sdk-react-native-test sdk-android-test sdk-ios-test sdk-capacitor-test

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

ci-example: build
	cd examples/ci && MOBILELAB_BIN=../../$(BINARY) ./run.sh

plugin-example:
	$(GO) build -trimpath -o $(PLUGIN_EXAMPLE_BINARY) ./examples/plugins/echo

plugin-example-check: build plugin-example
	cd examples/plugins && ../../$(BINARY) plugin list
	cd examples/plugins && ../../$(BINARY) plugin inspect echo --json
	cd examples/plugins && ../../$(BINARY) plugin run echo echo --input input.json

sdk-test: sdk-flutter-test sdk-react-native-test sdk-android-test sdk-ios-test sdk-capacitor-test

sdk-flutter-test:
	cd sdk/flutter/mobilelab_flutter && flutter pub get && dart format --output=none --set-exit-if-changed lib test example/lib && flutter analyze && flutter test

sdk-react-native-test:
	cd sdk/react-native && npm ci && npm run check && npm test && npm pack --dry-run

sdk-android-test:
	cd sdk/android/mobilelab-android && ./gradlew --no-daemon test jar

sdk-ios-test:
	cd sdk/ios/MobileLabKit && swift test

sdk-capacitor-test:
	cd sdk/capacitor && npm ci && npm run check && npm test && npm pack --dry-run

clean:
	rm -rf bin coverage.out
