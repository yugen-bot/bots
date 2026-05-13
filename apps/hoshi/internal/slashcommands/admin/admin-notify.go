package slashcommands

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
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

func (m *AdminNotifyModule) notify(ctx *discordgoplus.Ctx) {
	required := true
	err := discordgoplus.ModalRespond(ctx, &discordgo.InteractionResponseData{
		CustomID: "ADMIN_NOTIFY_SEND",
		Title:    "Send notification to all guilds",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID: "message",
						Label:    "Message to send",
						Style:    discordgo.TextInputParagraph,
						Required: &required,
					},
				},
			},
		},
	})
	if err != nil {
		utils.Logger.Error(err)
	}
}

func (m *AdminNotifyModule) handleNotifyModal(ctx *discordgoplus.Ctx) {
	fields := discordgoplus.ParseModalData(ctx.ModalData)
	content := fields["message"]

	ctx.Session.InteractionRespond(ctx.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	total, successByBotChannel, successByStarboard, err := m.notifyService.SendNotification(context.Background(), content)
	if err != nil {
		utils.Logger.Error(err)
		ctx.Session.FollowupMessageCreate(ctx.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong sending the notification.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	ctx.Session.FollowupMessageCreate(ctx.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Message sent to %d of %d guilds. %d were by `botUpdatesChannelId` settings.",
			successByBotChannel+successByStarboard, total, successByBotChannel),
		Flags: discordgo.MessageFlagsEphemeral,
	})
}

func (m *AdminNotifyModule) Modals() []*discordgoplus.Modal {
	return []*discordgoplus.Modal{
		{
			CustomID: "ADMIN_NOTIFY_SEND",
			Handler:  discordgoplus.HandlerFunc(m.handleNotifyModal),
		},
	}
}

func (m *AdminNotifyModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "notify",
			Description: "Send a notification to the configured channel of the bot",
			Handler:     discordgoplus.HandlerFunc(m.notify),
		},
	}
}
