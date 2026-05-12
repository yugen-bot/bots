package inits

import (
	"github.com/FedorLap2006/disgolf"
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	admin "jurien.dev/yugen/hoshi/internal/slashcommands/admin"
)

func RegisterModalHandlers(bot *disgolf.Bot, container *di.Container) {
	notifyModule := admin.GetAdminNotifyModule(container)

	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionModalSubmit {
			return
		}
		switch i.ModalSubmitData().CustomID {
		case "ADMIN_NOTIFY_SEND":
			notifyModule.HandleModal(s, i)
		}
	})
}
