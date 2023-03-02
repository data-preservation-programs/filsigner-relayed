lint:
	golangci-lint run

build:
	go build -o filsigner ./cmd
