VERSION=0.0.13
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"

all: check_ftp2

.PHONY: check_ftp2

check_ftp2: main.go
	go build $(LDFLAGS) -o check_ftp2

linux: main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o check_ftp2

check:
	go test -v ./...
	go test -race ./...

integration-test:
	./integration_test.sh

lint:
	golangci-lint run ./...