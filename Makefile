lint:
	golangci-lint run

build:
	go build -o filsigner ./cmd

gentypes:
	go generate ./model/response.go
