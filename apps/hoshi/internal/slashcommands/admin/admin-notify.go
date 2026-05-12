package slashcommands

import (
	"fmt"

	"github.com/FedorLap2006/disgolf"
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/utils"
)

type AdminNotifyModule struct {
	container     *di.Container
	notifyService *services.NotifyService
}

func GetAdminNotifyModule(container *di.Container) *AdminNotifyModule {
	return &AdminNotifyModule{
		container:     container,
		notifyService: container.Get(localStatic.DiNotify).(*services.NotifyService),
	}
}

func (m *AdminNotifyModule) notify(ctx *disgolf.Ctx) {
	err := ctx.Respond(&discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "ADMIN_NOTIFY_SEND",
			Title:    "Send notification to all guilds",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:  "message",
							Label:     "Message to send",
							Style:     discordgo.TextInputParagraph,
							Required:  true,
						},
					},
				},
			},
		},
	})
	if err != nil {
		utils.Logger.Error(err)
	}
}

func (m *AdminNotifyModule) HandleModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()

	content := ""
	for _, row := range data.Components {
		ar, ok := row.(*discordgo.ActionsRow)
		if !ok {
			continue
		}
		for _, comp := range ar.Components {
			if ti, ok := comp.(*discordgo.TextInput); ok && ti.CustomID == "message" {
				content = ti.Value
			}
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	total, successByBotChannel, successByStarboard, err := m.notifyService.SendNotification(content)
	if err != nil {
		utils.Logger.Error(err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong sending the notification.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Message sent to %d of %d guilds. %d were by `botUpdatesChannelId` settings.",
			successByBotChannel+successByStarboard, total, successByBotChannel),
		Flags: discordgo.MessageFlagsEphemeral,
	})
}

func (m *AdminNotifyModule) Commands() []*disgolf.Command {
	return []*disgolf.Command{
		{
			Name:        "notify",
			Description: "Send a notification to the configured channel of the bot",
			Handler:     disgolf.HandlerFunc(m.notify),
		},
	}
}
