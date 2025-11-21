build:
	@echo "Building Taco"
	@go build -o taco ./cmd/taco

install:
	@echo "Installing Taco"
	@go install ./cmd/taco
lint:
	@echo "Linting source code"
	@golangci-lint run