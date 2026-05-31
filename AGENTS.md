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

## Ent ORM — Code Generation

Each bot uses [entgo.io/ent](https://entgo.io) for its ORM. After modifying any file under `apps/<bot>/internal/ent/schema/`, regenerate the ent code from the **bot's root directory**:

```bash
cd apps/<bot>
GOWORK=off go generate ./internal/ent
```

**Always run this after schema changes.** The generator rewrites `runtime.go`, `mutation.go`, `settings.go`, `settings_create.go`, `settings_update.go`, `settings/settings.go`, `settings/where.go`, and `migrate/schema.go`. Editing those generated files by hand and skipping generation leaves `runtime.go` field indices stale, causing a runtime panic in `init()`.

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

# Run linter (also available as: make lint)
golangci-lint run

# Run the bot
make run-bot

# Run pocketbase
make run-pocketbase
```

## Code Style Guidelines

1. **Error Handling**: Always handle errors explicitly. Do not ignore errors with `_`. Wrap external errors: `fmt.Errorf("context: %w", err)`.
2. **Formatting**: Run `golangci-lint run --fix` (or `make lint`) after every edit. The linter enforces `gofumpt`, `goimports`, and `golines` (80-col max).
3. **Imports**: Three groups separated by blank lines — stdlib / `jurien.dev/yugen/*` / third-party. `goimports` with `-local jurien.dev/yugen` enforces this automatically.
4. **Naming**:
   - Use `ID` (not `Id`) for Discord/entity identifiers: `GuildID`, `ChannelID`, `GetByGuildID`.
   - Service method receivers: single letter (`s *SettingsService`, not `service`).
   - File names: kebab-case (`settings-show.go`, `embed-footer.go`).
5. **Slash command structure**: Each command group lives in its own sub-package under `slashcommands/`. Example:
   ```
   slashcommands/
     settings/
       settings.go       // struct, GetModule, Commands() sub-router wiring
       show/
         show.go         // package doc, struct, GetXxxModule, Commands()
         command.go      // handler functions
       set-channel/
         set-channel.go
         command.go
   ```
6. **Capability interfaces** (in `shared/utils/register-commands-module.go`): modules opt in via `Commands()`, `MessageComponents()`, and `Modals()` — never return empty slices for capabilities the command doesn't use.
7. **Listeners**: Use the struct-based form. Embed the session and bot in a struct; register handlers via methods.
8. **DI constants**: Every service exposes `Di<Name>` in `internal/static/di.go`; constructors accept only `*di.Container`.
9. **Early Returns**: Prefer early returns over nested if/else. Handle error cases first.
10. **Linting**: Run `make lint` before committing. Fix all lint errors — do not use `//nolint` without a comment explaining why.

## Testing

- Tests are in `*_test.go` files alongside source code
- Run `go test ./...` before committing changes

## Security Considerations

- Secrets & api tokens should always be store in the .env file, update the .env.example file to match it, but keep values blank!
- Always validate user input using validator tags
