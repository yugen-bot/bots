# Coding Agent Instructions

This document provides instructions for AI coding agents working on this repository.

## Project Overview

This is a **Golang/disgo** monorepo containing multiple applications in the apps folder. Each application has it's own purpose.

Here is a small summary of the applications:

`apps/hoshi` - A bot that handles starboards in discord servers
`apps/iro` - A bot that responds to hex color codes and represents the color
`apps/kazu` - A collaborative counting bot
`apps/kusari` - A collaborative word-chain bot
`apps/koto` - A collaborative word-guessing bot (Wordle-style)

## Technology Stack

- **Go**: 1.26+ required
- **Discord**: [disgoorg/disgo](https://github.com/disgoorg/disgo) v0.19+ via the [disgoplus](https://github.com/jurienhamaker/disgoplus) wrapper
- **API Framework**: [Gofiber](https://github.com/gofiber/fiber) v2.52+
- **ORM**: [entgo.io/ent](https://entgo.io)

## Directory Structure

```
.
├── go.work            # Go workspace for all modules
├── .golangci.yml      # Go linter configuration
│
├── apps/
│   └── [bot name]            # A named folder for each bot
│       ├── cmd                   # Entry points for the bot
│       ├── internal              # Per bot functionality
│       │   ├── inits                 # Initialization functions (DI, discord, commands)
│       │   ├── listeners             # Discord event listeners
│       │   ├── services              # Services that communicate with data
│       │   ├── slashcommands         # Slash commands
│       │   ├── static                # Static constants etc
│       │   └── utils                 # Utility functions
│       └── ent                   # Ent ORM schema + generated files
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
> Do **not** scaffold empty `services/`, `slashcommands/`, `static/`, `utils/`,
> or `ent/` directories for iro unless a feature genuinely requires them.

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

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Format code
go fmt ./...

# Run linter (also available as: make lint)
golangci-lint run
```

## Code Style Guidelines

1. **Error Handling**: Always handle errors explicitly. Do not ignore errors with `_`. Wrap external errors: `fmt.Errorf("context: %w", err)`.
2. **Formatting**: Run `golangci-lint run --fix` (or `make lint`) after every edit. The linter enforces `gofumpt`, `goimports`, and `golines` (80-col max).
3. **Imports**: Three groups separated by blank lines — stdlib / `jurien.dev/yugen/*` / third-party. `goimports` with `-local jurien.dev/yugen` enforces this automatically.
4. **Naming**:
   - Use `ID` (not `Id`) for Discord/entity identifiers: `GuildID`, `ChannelID`, `GetByGuildID`.
   - Service method receivers: single letter (`s *SettingsService`, not `service`).
   - File names: kebab-case (`settings-show.go`, `embed-footer.go`).
5. **Slash command structure (MANDATORY)**: Every leaf sub-command MUST live in its own sub-package (its own directory + `package`). This applies even when there is only one file inside.
   ```
   slashcommands/
     <group>/
       <group>.go        // package <group>; root command, sub-router wiring,
                         // capability fan-out for Modals/MessageComponents
       <leaf>/
         <leaf>.go       // package <leafpkg>; package doc, struct,
                         // Get<Leaf>Module(*di.Container), Commands(),
                         // and any capability methods this leaf opts into
         command.go      // handler func bodies
         embeds.go       // only when the leaf owns ≥ 1 embed builder
         modals.go       // only when the leaf implements Modals()
         handlers.go     // only when command.go would exceed ~2 handlers
         models.go       // only for local DTOs shared by ≥ 2 files
   ```
   Package identifiers strip hyphens from the directory name — `set-channel/` uses `package setchannel`, `start-after-first-guess/` uses `package startafterfirstguess`. Factory functions follow `Get<PascalCaseLeaf>Module(*di.Container)`, e.g. `setchannel.GetSetChannelModule(container)`.
   Group root files aggregate leaf sub-modules via a local `interface { Commands() []*disgoplus.Command }` loop (see `kazu/settings/settings.go` for the pattern).
   Single-file groups where the file is simultaneously the group root and the only command (`game/game.go` in kazu/kusari) are the sole exception — leave as-is until a sibling command is added.
6. **Capability interfaces** (in `shared/utils/register-commands-module.go`): modules opt in via `Commands()`, `MessageComponents()`, and `Modals()` — never return empty slices for capabilities the command doesn't use. Group-level files fan out to leaf capability methods rather than implementing them directly.
7. **Listeners**: Use the struct-based form with `*bot.Client`; register handlers via `client.EventManager.AddEventListeners(bot.NewListenerFunc(...))`.
8. **DI constants**: Every service exposes `Di<Name>` in `internal/static/di.go`; constructors accept only `*di.Container`. The bot is registered under `static.DiBot` as `*disgoplus.Bot`; retrieve with `container.Get(static.DiBot).(*disgoplus.Bot)`.
9. **Early Returns**: Prefer early returns over nested if/else. Handle error cases first.
10. **Linting**: Run `make lint` before committing. Fix all lint errors — do not use `//nolint` without a comment explaining why.

## disgoplus / disgo patterns

- **Bot creation** (`internal/inits/di.go`): each app creates `*disgoplus.Bot` via `disgoplus.New(token, sharded, gatewayOpt, cacheOpt, loggerOpt)`. Intents, presence, cache flags, and the logger are passed at construction time.
- **Presence**: set via `gateway.WithPresenceOpts(gateway.WithWatchingActivity("..."))` inside `bot.WithGatewayConfigOpts(...)` — not via `client.SetPresence` in event handlers.
- **Sharding**: when `cfg.Shard == true` use `bot.WithShardManagerConfigOpts(sharding.WithGatewayConfigOpts(...))` instead of `bot.WithGatewayConfigOpts(...)`.
- **Cache**: always pass `bot.WithCacheConfigOpts(cache.WithCaches(static.DefaultCacheFlags))`. Add `cache.FlagMessages` for bots that need message edit/delete history (hoshi, kazu, kusari).
- **Slash command handlers**: receive `*disgoplus.Ctx`. Options via `ctx.CommandData.Int("name")`, `ctx.CommandData.String("name")`, etc. Guild ID: `ctx.GuildID.String()`. Member: `ctx.Member.User.ID.String()`.
- **Responses**: `disgoplus.Defer(ctx, ephemeral)`, `disgoplus.FollowUp(ctx, discord.MessageCreate{...})`, `disgoplus.Update(ctx, discord.MessageUpdate{...})`, `disgoplus.ModalRespond(ctx, discord.ModalCreate{...})`.
- **REST calls**: `ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{...})`, `.AddReaction(...)`, `.GetGuild(...)`, etc.
- **Embeds**: `discord.NewEmbed().WithColor(...).WithTitle(...).WithDescription(...).WithEmbedFooter(footer)` — value type, no pointer.
- **Buttons**: `discord.NewPrimaryButton(label, customID)`, `discord.NewDangerButton(...)`, `discord.NewSecondaryButton(...)`. Action rows: `discord.NewActionRow(buttons...)`.
- **Custom-ID routing**: message components use slug params (`LEADERBOARD/:page`) parsed by disgoplus into `ctx.MessageComponentOptions`.

## Testing

- Tests are in `*_test.go` files alongside source code
- Run `go test ./...` before committing changes

## Security Considerations

- Secrets & api tokens should always be store in the .env file, update the .env.example file to match it, but keep values blank!
- Always validate user input using validator tags
