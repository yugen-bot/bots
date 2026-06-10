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

const errSomethingWentWrong = "Something went wrong, try again later."

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
		return fmt.Errorf("defer create message: %w", err)
	}

	page := 1
	if v, ok := data.OptInt("page"); ok {
		page = v
	}

	guildID := *e.GuildID()

	items, total, err := getItems(guildID, page)
	if err != nil {
		Logger.Errorw(
			"leaderboard: get items failed",
			"error",
			err,
			"guildID",
			guildID,
		)

		return sendCommandFollowup(e, errSomethingWentWrong)
	}

	if total == 0 {
		return sendCommandFollowup(
			e,
			"There is no leaderboard available yet for this server.",
		)
	}

	if len(items) == 0 {
		return sendCommandFollowup(
			e,
			fmt.Sprintf("No players found for page %d", page),
		)
	}

	embedColor := container.Get(static.DiEmbedColor).(int)
	cfg := container.Get(static.DiConfig).(*config.Config)
	bot := container.Get(static.DiBot).(*disgoplus.Bot)

	maxPage := int(math.Ceil(float64(total) / 10))
	footer := buildLeaderboardFooter(bot, cfg.OwnerID, page, maxPage)

	guild, err := bot.Client().Rest.GetGuild(guildID, false)
	if err != nil || guild == nil {
		Logger.Errorw(
			"leaderboard: get guild failed",
			"error",
			err,
			"guildID",
			guildID,
		)

		return sendCommandFollowup(e, errSomethingWentWrong)
	}

	embed := buildLeaderboardEmbed(embedColor, guild, items, formatter, footer)
	components := buildLeaderboardButtons(page, maxPage)

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Embeds:     []discord.Embed{embed},
		Components: components,
		Flags:      discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("create followup message: %w", err)
	}

	return nil
}

func LeaderboardComponentHandler(
	e *handler.ComponentEvent,
	container *di.Container,
	getItems LeaderboardDataFunc,
	formatter LeaderboardFormatFunc,
) error {
	page, err := strconv.Atoi(e.Vars["page"])
	if err != nil {
		return fmt.Errorf("parse page: %w", err)
	}

	guildID := *e.GuildID()

	items, total, err := getItems(guildID, page)
	if err != nil {
		Logger.Errorw(
			"leaderboard: get items failed",
			"error",
			err,
			"guildID",
			guildID,
		)

		return sendComponentUpdate(e, errSomethingWentWrong)
	}

	if total == 0 {
		return sendComponentUpdate(
			e,
			"There is no leaderboard available yet for this server.",
		)
	}

	if len(items) == 0 {
		return sendComponentUpdate(
			e,
			fmt.Sprintf("No players found for page %d", page),
		)
	}

	embedColor := container.Get(static.DiEmbedColor).(int)
	cfg := container.Get(static.DiConfig).(*config.Config)
	bot := container.Get(static.DiBot).(*disgoplus.Bot)

	maxPage := int(math.Ceil(float64(total) / 10))
	footer := buildLeaderboardFooter(bot, cfg.OwnerID, page, maxPage)

	guild, err := bot.Client().Rest.GetGuild(guildID, false)
	if err != nil || guild == nil {
		Logger.Errorw(
			"leaderboard: get guild failed",
			"error",
			err,
			"guildID",
			guildID,
		)

		return sendComponentUpdate(e, errSomethingWentWrong)
	}

	embed := buildLeaderboardEmbed(embedColor, guild, items, formatter, footer)
	components := buildLeaderboardButtons(page, maxPage)

	embeds := []discord.Embed{embed}

	if err := e.Respond(
		discord.InteractionResponseTypeUpdateMessage,
		discord.MessageUpdate{
			Embeds:     &embeds,
			Components: &components,
		},
	); err != nil {
		return fmt.Errorf("respond with update message: %w", err)
	}

	return nil
}

// sendCommandFollowup sends an ephemeral text followup and wraps any error.
func sendCommandFollowup(e *handler.CommandEvent, content string) error {
	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("create followup message: %w", err)
	}

	return nil
}

// sendComponentUpdate replaces the component message with a plain text message
// and wraps any error.
func sendComponentUpdate(e *handler.ComponentEvent, content string) error {
	empty := []discord.Embed{}
	emptyC := []discord.LayoutComponent{}

	if err := e.UpdateMessage(discord.MessageUpdate{
		Content:    &content,
		Embeds:     &empty,
		Components: &emptyC,
	}); err != nil {
		return fmt.Errorf("update message: %w", err)
	}

	return nil
}

func buildLeaderboardFooter(
	bot *disgoplus.Bot,
	ownerID string,
	page, maxPage int,
) *discord.EmbedFooter {
	footerParams := CreateEmbedFooterParams{IsVote: false}
	if maxPage > 1 {
		footerParams.Text = fmt.Sprintf("Page %d/%d", page, maxPage)
	}

	return CreateEmbedFooter(bot, &footerParams, ownerID)
}

func buildLeaderboardEmbed(
	embedColor int,
	guild *discord.RestGuild,
	items []any,
	formatter LeaderboardFormatFunc,
	footer *discord.EmbedFooter,
) discord.Embed {
	description := ""
	for i, item := range items {
		description = fmt.Sprintf(
			"%s\n%d. %s",
			description,
			i+1,
			formatter(item),
		)
	}

	thumbnailURL := ""
	if url := guild.IconURL(); url != nil {
		thumbnailURL = *url
	}

	return discord.NewEmbed().
		WithColor(embedColor).
		WithTitle(fmt.Sprintf("Points leaderboard for %s", guild.Name)).
		WithDescription(description).
		WithThumbnail(thumbnailURL).
		WithEmbedFooter(footer)
}

func buildLeaderboardButtons(page, maxPage int) []discord.LayoutComponent {
	var buttons []discord.InteractiveComponent

	if page > 1 {
		buttons = append(
			buttons,
			discord.NewPrimaryButton(
				"◀️",
				fmt.Sprintf("/LEADERBOARD/%d", page-1),
			),
		)
	}

	if page < maxPage {
		buttons = append(
			buttons,
			discord.NewPrimaryButton(
				"▶️",
				fmt.Sprintf("/LEADERBOARD/%d", page+1),
			),
		)
	}

	if len(buttons) == 0 {
		return nil
	}

	return []discord.LayoutComponent{discord.NewActionRow(buttons...)}
}
