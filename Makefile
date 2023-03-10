lint:
	golangci-lint run

build:
	go build -o filsigner ./cmd/filsigner

gentypes:
	go generate ./model/response.go
