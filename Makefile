.PHONY: build test clean

build: 
	go clean -testcache
	GOOS=linux GOARCH=amd64 go build -buildvcs=false -o=cmd/server ./cmd/server/...
	GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags "-X main.buildVersion=v1.0.1 -X main.buildDate=$(shell date +'%Y-%m-%d') -X main.buildCommit=$(shell git rev-parse HEAD)" -o=cmd/agent ./cmd/agent/...

test: build
	GOOS=linux GOARCH=amd64 go build -buildvcs=false -o=cmd/staticlint ./cmd/staticlint/...
	cmd/staticlint/staticlint ./...
	go mod tidy
	go clean -modcache
	go test ./... -coverprofile cover.out

cover: test
	go tool cover -html=cover.out -o coverage.html
	firefox coverage.html &
