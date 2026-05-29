# Simple Makefile for a Go project

all: build test

build:
	@echo "Building..."
	@go build -o dist/kazu cmd/bot/main.go

build-prod:
	@echo "Building..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o dist/kazu cmd/bot/main.go

ent-generate:
	@GOWORK=off go generate ./internal/ent

migrate:
	@atlas migrate apply --env local

migrate-diff:
	@atlas migrate diff $(name) --env local

migrate-validate:
	@atlas migrate validate --env local

run:
	@go run cmd/bot/main.go

test:
	@echo "Testing..."
	@go test ./... -v

clean:
	@echo "Cleaning..."
	@rm -f main

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

.PHONY: all build build-prod ent-generate migrate migrate-diff migrate-validate run test clean watch
