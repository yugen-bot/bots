package static

import "github.com/disgoorg/disgo/discord"

var (
	ButtonDiscordSupportServer = discord.NewLinkButton(
		"Join support server 👨‍⚕️",
		"https://support.yugen.bot",
	)
	ButtonKofi = discord.NewLinkButton(
		"Open Ko-Fi page ☕",
		"https://kofi.yugen.bot",
	)
)
