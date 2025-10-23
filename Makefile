# Cross-platform command definitions
ifeq ($(OS),Windows_NT)
	OPEN := start
	RM_CMD := if exist build rmdir /s /q build
	MKDIR_CMD := if not exist build mkdir build
else
	OS_NAME := $(shell uname)
	ifeq ($(OS_NAME), Darwin)
		OPEN := open
	else
		OPEN := xdg-open
	endif
	RM_CMD := rm -rf build/
	MKDIR_CMD := mkdir -p build
endif

BUILD_VERSION ?= "unknown"

clean:
ifeq ($(OS),Windows_NT)
	@$(RM_CMD)
else
	@rm -rf build/
endif

build: clean
ifeq ($(OS),Windows_NT)
	@$(MKDIR_CMD)
	@set GOOS=windows&& set GOARCH=amd64&& go build -o ./build/patreon-crawler.exe -ldflags "-X main.version=$(BUILD_VERSION)" ./main.go
	@set GOOS=linux&& set GOARCH=amd64&& go build -o ./build/patreon-crawler -ldflags "-X main.version=$(BUILD_VERSION)" ./main.go
else
	@GOOS=windows GOARCH=amd64 go build -o ./build/patreon-crawler.exe -ldflags "-X main.version=$(BUILD_VERSION)" ./main.go
	@GOOS=linux GOARCH=amd64 go build -o ./build/patreon-crawler -ldflags "-X main.version=$(BUILD_VERSION)" ./main.go
endif

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
