.PHONY: run fmt lint test mocks test-concurrent

run:
	go run cmd/server/main.go

fmt:
	go fmt ./...

lint:
	golangci-lint run

test:
	go test -v ./...

test-concurrent:
	@echo "Running concurrent load tests with k6..."
	@echo "\n=== Running Accuracy Test ==="
	@cd test/concurrent && k6 run accuracy_test.js

mocks:
	@echo "Generating mocks..."
	@mkdir -p modules/core/internal/test/mocks
	@echo "Generating mock for repository interfaces..."
	mockgen -source=modules/core/service/repository.go -destination=modules/core/internal/test/mocks/mock_repository.go -package=mocks
	@echo "Mocks generated successfully"
