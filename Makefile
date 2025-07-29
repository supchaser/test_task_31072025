run:
	go run cmd/main/main.go

TEST_PKGS := $(shell go list ./... | \
    grep -v /mocks)

run_tests:
	@echo "==> Running tests..."
	@go test $(GOFLAGS) -coverprofile coverage_raw.out -v $(TEST_PKGS)

test: run_tests
	@echo "==> Calculating coverage..."
	@grep -v "mock" coverage_raw.out > coverage.out
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o=coverage.html
	@echo "==> Done! Check coverage.html file!"
