# Coding Agent Instructions

This document provides instructions for AI coding agents working on this repository.

## Project Overview

This is a **Golang/Discordgo/Prisma** monorepo containing multiple applications in the apps folder. Each application has it's own purpose.

Here is a small summary of the applications:

`apps/hoshi` - A bot that handles starboards in discord servers
`apps/iro` - A bot that responds to hex color codes and represents the color
`apps/kazu` - A collaborative counting bot
`apps/kusari` - A collaborative word-chain bot

## Technology Stack

- **Go**: 1.25+ required
- Discord: [Discordgo](https://github.com/bwmarrin/discordgo) v0.27+ & [Discordgo-Plus](https://github.com/jurienhamaker/discordgo-plus)
- **API Framework**: [Gofiber](https://github.com/gofiber/fiber) v2.52+
- **ORM**: [prisma-client-go](https://github.com/steebchen/prisma-client-go)

## Directory Structure

```
.
├── go.mod             # Go mod file for all applications
├── .golangci.yml      # Go linter configuration
│
├── apps/
│   └── [bot name]            # A named folder for each bot
│       ├── cmd                   # Entry points for the bot
│       ├── internal              # Per bot functionality
│       │   ├── api                   # Api routes, middlewares, etc
│       │   ├── inits                 # Initialization functions
│       │   ├── listeners             # Discord listeners
│       │   ├── services              # Services that communicate with data
│       │   ├── slashcommands         # Slash commands
│       │   ├── static                # Static constants etc
│       │   └── utils                 # Utility functions
│       └── prisma                # Prisma files
│
├── shared/            # Shared code between applications
│   ├── api            # Api routes, middlewares, etc
│   ├── config         # Configuration files
│   ├── inits          # Initialization functions
│   ├── listeners      # Discord listeners
│   ├── metrics        # Prometheus metrics
│   ├── middlewares    # Shared middlewares
│   ├── slashcommands  # Shared slash commands
│   ├── static         # Static constants etc
│   └── utils          # Utility functions
│
└── assets             # Static assets for github etc
```

> **Exception — `apps/iro`**: Iro only responds to hex color codes in chat messages.
> It has no slash commands, no database, and no API routes.
> Its `internal/` directory intentionally contains only `inits/` and `listeners/`.
> Do **not** scaffold empty `api/`, `services/`, `slashcommands/`, `static/`, `utils/`,
> or `prisma/` directories for iro unless a feature genuinely requires them.

## Development Commands

```bash
# Install dependencies
go mod download

# Build the bot
make build-bot

# Build pocketbase
make build-pocketbase

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run the bot
make run-bot

# Run pocketbase
make run-pocketbase
```

## Code Style Guidelines

1. **Error Handling**: Always handle errors explicitly. Do not ignore errors with `_`.
2. **Style**:
   - Adhere to the rules in the .golangci.yml
   - Make sure all golangci errors are fixed
   - Use best practices as stated in [Effective Go](https://go.dev/doc/effective_go)
3. **Early Returns**: Prefer early returns over nested if/else blocks. Handle error cases and edge cases first, then proceed with the happy path. This reduces nesting and improves readability.
4. **Validation Tags**: Use `required` instead of deprecated `exists` tag

## Testing

- Tests are in `*_test.go` files alongside source code
- Run `go test ./...` before committing changes

## Security Considerations

- Secrets & api tokens should always be store in the .env file, update the .env.example file to match it, but keep values blank!
- Always validate user input using validator tags
