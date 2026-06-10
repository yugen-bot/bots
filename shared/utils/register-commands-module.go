package utils

import (
	"sync/atomic"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
)

var totalRegisteredCommands atomic.Int64

// TotalRegisteredCommands returns the number of leaf slash commands registered
// by the most recent RegisterCommandModules call.
func TotalRegisteredCommands() int {
	return int(totalRegisteredCommands.Load())
}

// CountLeafCommands counts the leaf-level commands across all modules —
// subcommands (and subcommands within groups) are each counted as 1;
// top-level commands with no sub-commands are counted as 1.
func CountLeafCommands(modules []RoutableModule) int {
	n := 0

	for _, m := range modules {
		for _, cmd := range m.Commands() {
			n += countLeafCreate(cmd)
		}
	}

	return n
}

func countLeafCreate(cmd discord.ApplicationCommandCreate) int {
	slashCmd, ok := cmd.(discord.SlashCommandCreate)
	if !ok {
		return 1
	}

	return countLeafOptions(slashCmd.Options)
}

func countLeafOptions(opts []discord.ApplicationCommandOption) int {
	leaves := 0
	hasSubCmds := false

	for _, opt := range opts {
		switch opt.Type() {
		case discord.ApplicationCommandOptionTypeSubCommand:
			hasSubCmds = true
			leaves++
		case discord.ApplicationCommandOptionTypeSubCommandGroup:
			hasSubCmds = true
			group := opt.(discord.ApplicationCommandOptionSubCommandGroup)
			leaves += len(group.Options)
		}
	}

	if !hasSubCmds {
		return 1
	}

	return leaves
}

// RoutableModule is implemented by every top-level slash-command module.
// Commands returns the ApplicationCommandCreate descriptors to register/sync,
// and Register wires all handlers (slash commands, components, modals) onto r.
type RoutableModule interface {
	Commands() []discord.ApplicationCommandCreate
	Register(r handler.Router)
}

// RegisterCommandModules builds a single handler.Mux from all modules and
// registers it as an event listener on the bot's client.
func RegisterCommandModules(bot *disgoplus.Bot, modules []RoutableModule) {
	mux := handler.New()

	for _, m := range modules {
		cmds := m.Commands()
		cmdCount := len(cmds)

		cmdStr := "commands"
		if cmdCount == 1 {
			cmdStr = "command"
		}

		m.Register(mux)

		name := commandsModuleName(cmds)
		Logger.Infof("Registered %q module with %d %s", name, cmdCount, cmdStr)
	}

	totalRegisteredCommands.Store(int64(CountLeafCommands(modules)))
	bot.Client().AddEventListeners(mux)
}

// SyncCommands collects all ApplicationCommandCreate definitions and syncs
// them to the given guild (or globally if guildID is zero).
func SyncCommands(
	bot *disgoplus.Bot,
	modules []RoutableModule,
	guildID snowflake.ID,
) error {
	var cmds []discord.ApplicationCommandCreate
	for _, m := range modules {
		cmds = append(cmds, m.Commands()...)
	}

	var guildIDs []snowflake.ID
	if guildID != 0 {
		guildIDs = []snowflake.ID{guildID}
	}

	Logger.Infof("Syncing %d commands", len(cmds))

	return handler.SyncCommands(bot.Client(), cmds, guildIDs)
}

func commandsModuleName(cmds []discord.ApplicationCommandCreate) string {
	if len(cmds) == 0 {
		return "unknown"
	}

	switch c := cmds[0].(type) {
	case discord.SlashCommandCreate:
		return c.Name
	case discord.UserCommandCreate:
		return c.Name
	case discord.MessageCommandCreate:
		return c.Name
	default:
		_ = c
		return "unknown"
	}
}
