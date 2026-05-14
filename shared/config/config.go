package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config holds all environment-driven configuration for every yugen bot.
// Load it once at startup via Load() and register it in DI.
type Config struct {
	// Core
	Env      string `env:"ENV"             envDefault:"development"`
	LogLevel string `env:"YUGEN_LOG_LEVEL" envDefault:"info"`
	Debug    bool   `env:"YUGEN_DEBUG"     envDefault:"false"`

	// Discord
	DiscordToken            string `env:"DISCORD_TOKEN,required"`
	DiscordAppID            string `env:"DISCORD_APP_ID,required"`
	DiscordDevelopmentGuild string `env:"DISCORD_DEVELOPMENT_GUILD_ID"`

	// Owner
	OwnerIDs []string `env:"OWNER_IDS" envSeparator:","`

	// Command sync
	SyncCommands bool `env:"SYNC_COMMANDS" envDefault:"false"`

	// API server
	APIHost string `env:"API_LISTEN_HOST" envDefault:"0.0.0.0"`
	APIPort string `env:"API_LISTEN_PORT" envDefault:"8080"`

	// Loki
	LokiHost     string `env:"LOKI_HOST"`
	LokiUsername string `env:"LOKI_USERNAME"`
	LokiPassword string `env:"LOKI_PASSWORD"`

	// Report channels
	LogsChannelID string `env:"DISCORD_LOGS_REPORT_CHANNEL_ID"`
	VoteChannelID string `env:"DISCORD_VOTE_REPORT_CHANNEL_ID"`

	// Top.GG
	TopGGSync     bool   `env:"TOP_GG_SYNC"      envDefault:"false"`
	TopGGToken    string `env:"TOP_GG_TOKEN"`
	TopGGVoteLink string `env:"TOP_GG_VOTE_LINK"`

	// DiscordBotList
	DiscordBotListSync     bool   `env:"DISCORDBOTLIST_SYNC"      envDefault:"false"`
	DiscordBotListToken    string `env:"DISCORDBOTLIST_TOKEN"`
	DiscordBotListVoteLink string `env:"DISCORDBOTLIST_VOTE_LINK"`

	// Webhook auth
	WebhookAuthorizationToken string `env:"WEBHOOK_AUTHORIZATION_TOKEN"`

	// Invite link
	InviteLink string `env:"INVITE_LINK"`

	// Owner user ID (for embed footer; first entry of OwnerIDs is preferred,
	// but kept separately for backwards compatibility)
	OwnerID string `env:"OWNER_ID"`

	// Per-bot extras (kusari)
	WiktionaryUsername string `env:"WIKTIONARY_USERNAME"`
	WiktionaryPassword string `env:"WIKTIONARY_PASSWORD"`

	// Sentry
	SentryDSN              string  `env:"SENTRY_DSN"`
	SentryEnvironment      string  `env:"SENTRY_ENVIRONMENT"`
	SentryTracesSampleRate float64 `env:"SENTRY_TRACES_SAMPLE_RATE" envDefault:"0.0"`
	SentryDebug            bool    `env:"SENTRY_DEBUG"              envDefault:"false"`
}

// Load parses environment variables into a Config.
// Returns an error describing any missing required fields.
func Load() (Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return c, fmt.Errorf("config: load: %w", err)
	}
	// If OWNER_IDS is set, use the first entry as the primary OwnerID
	// so callers don't need to know about both fields.
	if c.OwnerID == "" && len(c.OwnerIDs) > 0 {
		c.OwnerID = c.OwnerIDs[0]
	}

	return c, nil
}
