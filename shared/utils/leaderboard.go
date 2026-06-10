package utils

import (
	"fmt"
	"math"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
)

type (
	LeaderboardDataFunc   func(guildID snowflake.ID, page int) ([]any, int, error)
	LeaderboardFormatFunc func(item any) string
)

func GetLeaderboardCommands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "leaderboard",
			Description: "Get the current servers leaderboard!",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "page",
					Description: "View a specific page.",
					Required:    false,
				},
			},
		},
	}
}

func LeaderboardCommandHandler(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
	container *di.Container,
	getItems LeaderboardDataFunc,
	formatter LeaderboardFormatFunc,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	page := 1
	if v, ok := data.OptInt("page"); ok {
		page = v
	}

	guildID := *e.GuildID()

	items, total, err := getItems(guildID, page)
	if err != nil {
		Logger.Errorw("leaderboard: get items failed", "error", err, "guildID", guildID)
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	if total == 0 {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "There is no leaderboard available yet for this server.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	if len(items) == 0 {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf("No players found for page %d", page),
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	embedColor := container.Get(static.DiEmbedColor).(int)
	cfg := container.Get(static.DiConfig).(*config.Config)
	bot := container.Get(static.DiBot).(*disgoplus.Bot)

	maxPage := int(math.Ceil(float64(total) / 10))

	footerParams := CreateEmbedFooterParams{IsVote: false}
	if maxPage > 1 {
		footerParams.Text = fmt.Sprintf("Page %d/%d", page, maxPage)
	}
	footer := CreateEmbedFooter(bot, &footerParams, cfg.OwnerID)

	guild, err := bot.Client().Rest.GetGuild(guildID, false)
	if err != nil || guild == nil {
		Logger.Errorw("leaderboard: get guild failed", "error", err, "guildID", guildID)
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	description := ""
	for i, item := range items {
		description = fmt.Sprintf("%s\n%d. %s", description, i+1, formatter(item))
	}

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle(fmt.Sprintf("Points leaderboard for %s", guild.Name)).
		WithDescription(description).
		WithThumbnail(func() string {
			if url := guild.IconURL(); url != nil {
				return *url
			}
			return ""
		}()).
		WithEmbedFooter(footer)

	var buttons []discord.InteractiveComponent
	if page > 1 {
		buttons = append(buttons, discord.NewPrimaryButton("◀️", fmt.Sprintf("/LEADERBOARD/%d", page-1)))
	}
	if page < maxPage {
		buttons = append(buttons, discord.NewPrimaryButton("▶️", fmt.Sprintf("/LEADERBOARD/%d", page+1)))
	}

	var components []discord.LayoutComponent
	if len(buttons) > 0 {
		components = []discord.LayoutComponent{discord.NewActionRow(buttons...)}
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Embeds:     []discord.Embed{embed},
		Components: components,
		Flags:      discord.MessageFlagEphemeral,
	})
	return err
}

func LeaderboardComponentHandler(
	e *handler.ComponentEvent,
	container *di.Container,
	getItems LeaderboardDataFunc,
	formatter LeaderboardFormatFunc,
) error {
	page, err := strconv.Atoi(e.Vars["page"])
	if err != nil {
		return nil
	}

	guildID := *e.GuildID()

	items, total, err := getItems(guildID, page)
	if err != nil {
		Logger.Errorw("leaderboard: get items failed", "error", err, "guildID", guildID)
		content := "Something went wrong, try again later."
		empty := []discord.Embed{}
		emptyC := []discord.LayoutComponent{}
		return e.UpdateMessage(discord.MessageUpdate{
			Content:    &content,
			Embeds:     &empty,
			Components: &emptyC,
		})
	}

	if total == 0 {
		content := "There is no leaderboard available yet for this server."
		empty := []discord.Embed{}
		emptyC := []discord.LayoutComponent{}
		return e.UpdateMessage(discord.MessageUpdate{
			Content:    &content,
			Embeds:     &empty,
			Components: &emptyC,
		})
	}

	if len(items) == 0 {
		content := fmt.Sprintf("No players found for page %d", page)
		empty := []discord.Embed{}
		emptyC := []discord.LayoutComponent{}
		return e.UpdateMessage(discord.MessageUpdate{
			Content:    &content,
			Embeds:     &empty,
			Components: &emptyC,
		})
	}

	embedColor := container.Get(static.DiEmbedColor).(int)
	cfg := container.Get(static.DiConfig).(*config.Config)
	bot := container.Get(static.DiBot).(*disgoplus.Bot)

	maxPage := int(math.Ceil(float64(total) / 10))

	footerParams := CreateEmbedFooterParams{IsVote: false}
	if maxPage > 1 {
		footerParams.Text = fmt.Sprintf("Page %d/%d", page, maxPage)
	}
	footer := CreateEmbedFooter(bot, &footerParams, cfg.OwnerID)

	guild, err := bot.Client().Rest.GetGuild(guildID, false)
	if err != nil || guild == nil {
		Logger.Errorw("leaderboard: get guild failed", "error", err, "guildID", guildID)
		content := "Something went wrong, try again later."
		empty := []discord.Embed{}
		emptyC := []discord.LayoutComponent{}
		return e.UpdateMessage(discord.MessageUpdate{
			Content:    &content,
			Embeds:     &empty,
			Components: &emptyC,
		})
	}

	description := ""
	for i, item := range items {
		description = fmt.Sprintf("%s\n%d. %s", description, i+1, formatter(item))
	}

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle(fmt.Sprintf("Points leaderboard for %s", guild.Name)).
		WithDescription(description).
		WithThumbnail(func() string {
			if url := guild.IconURL(); url != nil {
				return *url
			}
			return ""
		}()).
		WithEmbedFooter(footer)

	var buttons []discord.InteractiveComponent
	if page > 1 {
		buttons = append(buttons, discord.NewPrimaryButton("◀️", fmt.Sprintf("/LEADERBOARD/%d", page-1)))
	}
	if page < maxPage {
		buttons = append(buttons, discord.NewPrimaryButton("▶️", fmt.Sprintf("/LEADERBOARD/%d", page+1)))
	}

	var components []discord.LayoutComponent
	if len(buttons) > 0 {
		components = []discord.LayoutComponent{discord.NewActionRow(buttons...)}
	}

	embeds := []discord.Embed{embed}
	return e.Respond(discord.InteractionResponseTypeUpdateMessage, discord.MessageUpdate{
		Embeds:     &embeds,
		Components: &components,
	})
}
