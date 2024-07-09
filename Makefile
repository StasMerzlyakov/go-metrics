.PHONY: build test clean

build: 
	GOOS=linux GOARCH=amd64 go build -buildvcs=false -o=cmd/server ./cmd/server/...
	GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags "-X main.buildVersion=v1.0.1 -X main.buildDate=$(shell date +'%Y-%m-%d') -X main.buildCommit=$(shell git rev-parse HEAD)" -o=cmd/agent ./cmd/agent/...
	GOOS=linux GOARCH=amd64 go build -buildvcs=false -o=cmd/keygen ./cmd/keygen/...

test: build
	GOOS=linux GOARCH=amd64 go build -buildvcs=false -o=cmd/staticlint ./cmd/staticlint/...
	cmd/staticlint/staticlint ./...
	go mod tidy
	go clean -testcache
	go test ./... -coverprofile cover.out.tmp
	cat cover.out.tmp | grep -v "mocks" > cover.out
	rm cover.out.tmp

cover: test
	go tool cover -html=cover.out -o coverage.html
	firefox coverage.html &

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  internal/proto/metrics.proto 

