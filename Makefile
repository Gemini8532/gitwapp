.PHONY: all build-frontend build-backend nginx-config init-env run-prod check-env

# Default values for .env generation
PORT ?= 8084
APP_NAME ?= "My Go App"

# Check that .env exists
check-env:
	@test -f .env || (echo "Error: .env file not found. Run 'make init-env' first." && exit 1)

# Load .env if it exists
ifneq (,$(wildcard ./.env))  
    include .env  
    export  
endif

all: check-env build-frontend build-backend

init-env:
	@sed -e 's/{{PORT}}/$(PORT)/g' -e 's/{{APP_NAME}}/$(APP_NAME)/g' templates/env > .env
	@echo "Generated .env with APP_PORT=$(PORT) and VITE_APP_NAME=$(APP_NAME)"

build-frontend: check-env
	cd frontend && npm install && npm run build

# Track all Go files in cmd/server and internal for dependency tracking
SERVER_SOURCES := $(shell find cmd/server internal -name '*.go' -type f 2>/dev/null)

# Build information
VERSION ?= dev
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

build-backend: check-env $(SERVER_SOURCES)
	@mkdir -p bin
	@mkdir -p frontend/dist
	@touch frontend/dist/.keep
	go build -ldflags "\
		-X main.defaultPort=$(APP_PORT) \
		-X main.version=$(VERSION) \
		-X main.buildDate=$(BUILD_DATE) \
		-X main.gitCommit=$(GIT_COMMIT)" \
		-o bin/server ./cmd/server
	@echo "Built bin/server v$(VERSION) ($(GIT_COMMIT)) with port $(APP_PORT)"

nginx-config: check-env
	@sed 's/{{APP_PORT}}/$(APP_PORT)/g' templates/nginx.conf

run-prod: all  
	./bin/server serve
