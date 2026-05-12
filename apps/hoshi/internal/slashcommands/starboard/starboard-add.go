package slashcommands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	localUtils "jurien.dev/yugen/hoshi/internal/utils"
	"jurien.dev/yugen/shared/static"
)

type StarboardAddModule struct {
	container *di.Container
	starboard *services.StarboardService
}

func GetStarboardAddModule(container *di.Container) *StarboardAddModule {
	return &StarboardAddModule{
		container: container,
		starboard: container.Get(localStatic.DiStarboard).(*services.StarboardService),
	}
}

func (m *StarboardAddModule) add(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	bot := m.container.Get(static.DiBot).(*discordgoplus.Bot)
	destination := ctx.Options["destination"].ChannelValue(ctx.Session)

	emojiInput := "⭐"
	if opt, ok := ctx.Options["emoji"]; ok {
		emojiInput = opt.StringValue()
	}

	found, key, display, unicode := localUtils.ResolveEmoji(emojiInput, bot)
	if !found {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "You can only use emojis from guilds that the bot is in.",
		}, true)
		return
	}

	var sourceChannelID *string
	sourceLabel := ""
	if opt, ok := ctx.Options["source"]; ok {
		src := opt.ChannelValue(ctx.Session)
		id := src.ID
		sourceChannelID = &id
		sourceLabel = fmt.Sprintf("\nSource: <#%s>", id)
	}

	existing, _ := m.starboard.GetStarboardBySourceIDAndEmoji(ctx.Interaction.GuildID, key, sourceChannelID)
	if existing != nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "A starboard for the supplied rules already exists.",
		}, true)
		return
	}

	_, err := m.starboard.AddStarboard(ctx.Interaction.GuildID, key, sourceChannelID, destination.ID)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	emojiDisplay := display
	if unicode {
		emojiDisplay = key
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("A starboard has been added;\nDestination: <#%s>\nEmoji: %s%s",
			destination.ID, emojiDisplay, sourceLabel),
	}, true)
}

func (m *StarboardAddModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "add",
			Description: "Add a starboard",
			Handler:     discordgoplus.HandlerFunc(m.add),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "destination",
					Description: "The destination channel to keep the starboard in",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "emoji",
					Description: "An emoji to check for (default ⭐)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "source",
					Description: "A source channel to check",
					Required:    false,
				},
			},
		},
	}
}
