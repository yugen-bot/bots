package utils

import (
	"sync/atomic"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
)

var totalRegisteredCommands atomic.Int64

// TotalRegisteredCommands returns the number of leaf slash commands recorded
// by the most recent SetTotalRegisteredCommands call.
func TotalRegisteredCommands() int {
	return int(totalRegisteredCommands.Load())
}

// SetTotalRegisteredCommands stores the leaf-count gauge consumed by the
// discord_stat_total_interactions metric.
func SetTotalRegisteredCommands(n int) {
	totalRegisteredCommands.Store(int64(n))
}

// CountLeafCommands counts the leaf-level commands across all modules —
// subcommands (and subcommands within groups) are each counted as 1;
// top-level commands with no sub-commands are counted as 1.
func CountLeafCommands(modules []disgoplus.RoutableModule) int {
	n := 0

	for _, m := range modules {
		for _, reg := range m.Commands() {
			n += countLeafCreate(reg.Create)
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
