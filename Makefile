default: test

# -timeout	timout in seconds
#  -v		verbose output
test:
	@ echo "+ $@"
	@ go test -timeout=5s -v ./...

build:
	@ echo "+ $@"
	@ CGO_ENABLED=0 go build -v -installsuffix cgo -o ./bin/slag .

# Cross-compile
# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5
build-linux:
	@ echo "+ $@"
	@ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o ./bin/slag-linux-amd64 .

build-mac:
	@ echo "+ $@"
	@ GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o ./bin/slag-darwin-amd64 .

run: build
	@ echo "+ $@"
	@ ./bin/slag

install:
	@ echo "+ $@"
	@ go install .

build-all: build build-linux build-mac

fmt:
	go fmt ./...

.PHONY: default test build build-linux build-mac run install fmt
