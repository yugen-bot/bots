package static

import "github.com/disgoorg/disgo/discord"

var (
	ButtonDiscordSupportServer = discord.NewLinkButton(
		"Join support server 👨‍⚕️",
		"https://discord.gg/UttZbEd9zn",
	)
	ButtonKofi = discord.NewLinkButton(
		"Open Ko-Fi page ☕",
		"https://ko-fi.com/jurienhamaker",
	)
)
