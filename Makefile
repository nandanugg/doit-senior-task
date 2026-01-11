.PHONY: run fmt lint

run:
	go run main.go

fmt:
	go fmt ./...

lint:
	golangci-lint run
