package utils

// TODO: test requires mock
// SyncCommands requires a live *discordgoplus.Bot (which wraps a
// *discordgo.Session) and a *config.Config. Calling it without a real Discord
// connection would panic when bot.Router.Sync tries to make HTTP requests.
//
// The function contains no extractable pure sub-function; the only logic is a
// single cfg.SyncCommands flag guard. Coverage for this file should come from
// integration tests with a stub Bot implementation.
