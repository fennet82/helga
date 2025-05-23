
APP_NAME=helga
CONF_FILE_PATH=helga_conf_example.yaml
EXECUTABLE_PATH="./bin/helga"
SHELL := /bin/bash

.EXPORT_ALL_VARIABLES:
export GO111MODULE=on
export LOGS_FILE_PATH=./$(APP_NAME).log 
export HELGA_CONF_FILE_PATH=$(CONF_FILE_PATH)

.PHONY: all setup check-quality lint vet fmt tidy test-coverage build run vendor clean

all: setup check-quality test-coverage build run clean

setup: tidy vendor

vendor:
	@echo "--> Vendoring Go modules..."
	go mod vendor
	@echo "Vendor complete."

tidy:
	@echo "--> Tidying Go modules..."
	go mod tidy
	@echo "Tidy complete."

check-quality: lint fmt vet

lint:
	@echo "--> Running golangci-lint..."
	golangci-lint run || true
	@echo "Linting complete."

vet:
	@echo "--> Running go vet..."
	go vet ./...
	@echo "Vet complete."

fmt:
	@echo "--> Running go fmt..."
	go fmt ./...
	@echo "Fmt complete."

test-coverage:
	@echo "--> Running tests with coverage..."
	go test -v -timeout 10m ./... -coverprofile=coverage.out -json > report.json
	go tool cover -html=coverage.out
	@echo "Test coverage report generated."

build: setup
	@echo "--> Building $(APP_NAME) executable..."
	mkdir -p bin/
	go build -mod=vendor -o $(EXECUTABLE_PATH) ./cmd
	@echo "Build passed. Executable created at $(EXECUTABLE_PATH)"

run: build
	@echo "--> Running $(APP_NAME)..."
	chmod +x $(EXECUTABLE_PATH)
	$(EXECUTABLE_PATH)

clean:
	@echo "--> Cleaning up generated files..."
	go clean
	rm -rf bin/
	rm -rf vendor/
	rm -f cover*.out report.json $(APP_NAME).log
	@echo "Clean complete."
