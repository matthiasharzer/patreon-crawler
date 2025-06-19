OS_NAME := $(shell uname)
ifeq ($(OS_NAME), Darwin)
OPEN := open
else
OPEN := xdg-open
endif

BUILD_VERSION ?= "unknown"

clean:
	@rm -rf build/

build: clean
	@GOOS=windows GOARCH=amd64 go build -o ./build/patreon-crawler.exe -ldflags "-X main.version=$(BUILD_VERSION)" ./main.go
	@GOOS=linux GOARCH=amd64 go build -o ./build/patreon-crawler -ldflags "-X main.version=$(BUILD_VERSION)" ./main.go

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
