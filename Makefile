OS_NAME := $(shell uname)
ifeq ($(OS_NAME), Darwin)
OPEN := open
else
OPEN := xdg-open
endif

BUILD_VERSION ?= "unknown"

build:
	@rm -rf build/

	@GOOS=windows GOARCH=amd64 go build -o ./build/patreon-crawler.exe -ldflags "-X main.version=$(BUILD_VERSION)" ./main.go

	@GOOS=linux GOARCH=amd64 go build -o ./build/patreon-crawler -ldflags "-X main.version=$(BUILD_VERSION)" ./main.go

.PHONY: build
