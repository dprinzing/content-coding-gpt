.DEFAULT_GOAL := build

# Format Code
format:
	@echo "Formatting code:"
	go fmt ./...
.PHONY:format

# Check Code Style
# go install honnef.co/go/tools/cmd/staticcheck@latest
# go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
lint: format
	@echo "Linting code:"
	staticcheck ./...
	shadow ./...
	go vet ./...
.PHONY:lint

# Test Code
test: lint
	@echo "Testing code:"
	go test ./...
.PHONY:test

# Update Dependencies
dependencies:
	@echo "Updating dependencies:"
	go get -u ./...
	go mod tidy
.PHONY:dependencies

# Install/Update Tools
tools:
	@echo "Installing/updating tools:"
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
.PHONY:tools

# Build the command-line applications
build:
	@echo "Building gpt command for local use:"
	go build -o ./gpt ./cmd/*.go
.PHONY:build
