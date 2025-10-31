build:
	@echo "Building Taco"
	@go build -o taco ./cmd/taco

lint:
	@echo "Linting source code"
	@golangci-lint run