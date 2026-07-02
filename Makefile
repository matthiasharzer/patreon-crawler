BUILD_VERSION ?= "unknown"

OUTPUT_NAME := "patreon-crawler"
MODULE_NAME := $(shell go list -m)

clean:
	@rm -rf build/

build: clean
	@GOOS=windows GOARCH=amd64 go build -o ./build/$(OUTPUT_NAME)-windows-amd64.exe -ldflags "-X $(MODULE_NAME)/cmd/version.version=$(BUILD_VERSION)" ./main.go
	@GOOS=linux GOARCH=amd64 go build -o ./build/$(OUTPUT_NAME)-linux-amd64 -ldflags "-X $(MODULE_NAME)/cmd/version.version=$(BUILD_VERSION)" ./main.go
	@GOOS=linux GOARCH=arm64 go build -o ./build/$(OUTPUT_NAME)-linux-arm64 -ldflags "-X $(MODULE_NAME)/cmd/version.version=$(BUILD_VERSION)" ./main.go

qa: analyze test

analyze:
	@go vet
	@go run honnef.co/go/tools/cmd/staticcheck@latest --checks=all

test:
	@go test -failfast -cover ./...

.PHONY: build \
		analyze \
		qa \
		test \
		clean
