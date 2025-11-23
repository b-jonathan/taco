build:
	@echo "Building Taco"
	@go build -o taco ./cmd/taco

install:
	@echo "Installing Taco"
	@go install ./cmd/taco
lint:
	@echo "Linting source code"
	@golangci-lint run

setup:
	@echo "Setting up pre-commit hooks"
	@go install github.com/evilmartians/lefthook@latest
	@lefthook install

build-all:
	go build -o npm/bin/taco-windows-amd64.exe ./cmd/taco
	GOOS=darwin GOARCH=amd64 go build -o npm/bin/taco-darwin-amd64 ./cmd/taco
	GOOS=darwin GOARCH=arm64 go build -o npm/bin/taco-darwin-arm64 ./cmd/taco
	GOOS=linux GOARCH=amd64 go build -o npm/bin/taco-linux-amd64 ./cmd/taco
	GOOS=linux GOARCH=arm64 go build -o npm/bin/taco-linux-arm64 ./cmd/taco
