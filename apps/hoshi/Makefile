# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building..."
	@go build -o dist/hoshi cmd/bot/main.go

build-prod:
	@echo "Building..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o dist/hoshi cmd/bot/main.go

ent-generate:
	@GOWORK=off go generate ./internal/ent

migrate:
	@atlas migrate apply --env local

migrate-diff:
	@atlas migrate diff $(name) --env local

migrate-validate:
	@atlas migrate validate --env local

# Run the application
run:
	@go run cmd/bot/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test clean watch ent-generate migrate migrate-diff migrate-validate
