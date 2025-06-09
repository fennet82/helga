# CONSTS
COLOR_PURPLE=\\033[1;35m
COLOR_RESET=\\033[0m

# DEFAULT VARS (can be overriden by local env vars with same name)
CWD ?= $(shell pwd)
APP_NAME = helga
EXECUTABLE_PATH ?= $(CWD)/bin
GOLINT_FILE_PATH ?= $(CWD)/.golangci.yaml
SHELL ?= /bin/bash

# ENV VARS AVAILABLE TO ALL TARGETS
.EXPORT_ALL_VARIABLES:
GO111MODULE ?= on
LOGS_FILE_PATH ?= $(CWD)/$(APP_NAME).log
HELGA_CONF_FILE_PATH ?=  $(CWD)/helga_conf_example.yaml

# Main targets
.PHONY: all
all: setup check-quality test-coverage build run clean

.PHONY: setup
setup: tidy

.PHONY: vendor
vendor:
	@echo -e "${COLOR_PURPLE}--> Vendoring Go modules...${COLOR_RESET}"
	go mod vendor
	@echo -e "${COLOR_PURPLE}Vendor complete.${COLOR_RESET}"

.PHONY: tidy
tidy:
	@echo -e "${COLOR_PURPLE}--> Tidying Go modules...${COLOR_RESET}"
	go mod tidy
	@echo -e "${COLOR_PURPLE}Tidy complete.${COLOR_RESET}"

.PHONY: check-quality
check-quality: lint fmt vet

.PHONY: lint
lint:
	@echo -e "${COLOR_PURPLE}--> Running golangci-lint...${COLOR_RESET}"
	golangci-lint -c $(GOLINT_FILE_PATH) run || true
	@echo -e "${COLOR_PURPLE}Linting complete.${COLOR_RESET}"

.PHONY: vet
vet:
	@echo -e "${COLOR_PURPLE}--> Running go vet...${COLOR_RESET}"
	go vet ./...
	@echo -e "${COLOR_PURPLE}Vet complete.${COLOR_RESET}"

.PHONY: fmt
fmt:
	@echo -e "${COLOR_PURPLE}--> Running go fmt...${COLOR_RESET}"
	go fmt ./...
	@echo -e "${COLOR_PURPLE}Fmt complete.${COLOR_RESET}"

.PHONY: test
test:
	@echo -e "${COLOR_PURPLE}--> Running tests and creating coverage...${COLOR_RESET}"
	go test -v -timeout 10m ./... -coverprofile=coverage.out -json > report.json
	@echo -e "${COLOR_PURPLE}Tests completed and test coverage report generated.${COLOR_RESET}"

.PHONY: test-coverage
test-coverage: test
	@echo -e "${COLOR_PURPLE}--> creating coverage as html...${COLOR_RESET}"
	go tool cover -html=coverage.out
	@echo -e "${COLOR_PURPLE}Test coverage html report generated.${COLOR_RESET}"

.PHONY: build
build: setup
	@echo -e "${COLOR_PURPLE}--> Building $(APP_NAME) executable...${COLOR_RESET}"
	mkdir -p bin/
	go build -o $(EXECUTABLE_PATH)/$(APP_NAME) $(CWD)/cmd/main
	@echo -e "${COLOR_PURPLE}Build passed. Executable created at $(EXECUTABLE_PATH)/$(APP_NAME)${COLOR_RESET}"

.PHONY: run
run: build
	@echo -e "${COLOR_PURPLE}--> Running $(APP_NAME)...${COLOR_RESET}"
	chmod +x $(EXECUTABLE_PATH)/$(APP_NAME)
	$(EXECUTABLE_PATH)/$(APP_NAME)

.PHONY: clean
clean:
	@echo -e "${COLOR_PURPLE}--> Cleaning up generated files...${COLOR_RESET}"
	go clean
	rm -rf bin/
	rm -rf vendor/
	rm -f cover*.out report.json 
	rm -f $(APP_NAME).log
	@echo -e "${COLOR_PURPLE}Clean complete.${COLOR_RESET}"

# .PHONY: deploy
# deploy:

# Playground targets
.PHONY: pg-init
pg-init: setup
	@echo -e "${COLOR_PURPLE}--> Creating playground.go under ./playground${COLOR_RESET}"
	mkdir -p ./playground
	touch ./playground/main.go
	@echo -e "${COLOR_PURPLE}--> file created${COLOR_RESET}"

.PHONY: pg-build
pg-build: setup
	@echo -e "${COLOR_PURPLE}--> Building playground executable...${COLOR_RESET}"
	mkdir -p bin/
	go build -mod=vendor -o $(EXECUTABLE_PATH)/playground ./playground
	@echo -e "${COLOR_PURPLE}Build passed. Executable created at $(EXECUTABLE_PATH)/playground${COLOR_RESET}"

.PHONY: pg-run
pg-run: pg-build
	@echo -e "${COLOR_PURPLE}--> Running playground...${COLOR_RESET}"
	chmod +x $(EXECUTABLE_PATH)/playground
	$(EXECUTABLE_PATH)/playground

.PHONY: pg-clean
pg-clean:
	@echo -e "${COLOR_PURPLE}--> cleaning playground${COLOR_RESET}"
	rm -rf playground
	rm -f $(EXECUTABLE_PATH)/playground
	@echo -e "${COLOR_PURPLE}Clean complete.${COLOR_RESET}"
