GO ?= go
BINARY := bin/mobilelab
VERSION ?= 0.5.0
LDFLAGS ?= -s -w -X github.com/mobilelab-dev/mobilelab/internal/cli.Version=$(VERSION)

.PHONY: build test lint run clean release-check sdk-test sdk-flutter-test sdk-react-native-test sdk-android-test sdk-ios-test sdk-capacitor-test

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
