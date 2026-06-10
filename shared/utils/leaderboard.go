package utils

import (
	"fmt"
	"math"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
)

type leaderboardSourceType string

const (
	LEADERBOARD_INTERACTION       leaderboardSourceType = "interaction"
	LEADERBOARD_MESSAGE_COMPONENT leaderboardSourceType = "message_component"
)

type (
	LeaderboardDataFunc   func(ctx *disgoplus.Ctx, page int) ([]any, int, error)
	LeaderboardFormatFunc func(ctx *disgoplus.Ctx, item any) string
)

func GetLeaderboardCommands(handler disgoplus.HandlerFunc) []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "leaderboard",
			Description: "Get the current servers leaderboard!",
			Handler:     handler,
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

func GetLeaderboardMessageComponents(handler disgoplus.HandlerFunc) []*disgoplus.MessageComponent {
	return []*disgoplus.MessageComponent{
		{
			CustomID: "LEADERBOARD/:page",
			Handler:  handler,
		},
	}
}

func LeaderboardCommandHandler(
	ctx *disgoplus.Ctx,
	container *di.Container,
	getItems LeaderboardDataFunc,
	formatter LeaderboardFormatFunc,
) {
	page := 1
	if ctx.CommandData != nil {
		if v, ok := ctx.CommandData.OptInt("page"); ok {
			page = v
		}
	}

	ShowLeaderboard(ctx, container, LEADERBOARD_INTERACTION, page, getItems, formatter)
}

func LeaderboardMessageComponentHandler(
	ctx *disgoplus.Ctx,
	container *di.Container,
	getItems LeaderboardDataFunc,
	formatter LeaderboardFormatFunc,
) {
	page, err := strconv.Atoi(ctx.MessageComponentOptions["page"])
	if err != nil {
		return
	}

	ShowLeaderboard(ctx, container, LEADERBOARD_MESSAGE_COMPONENT, page, getItems, formatter)
}

func ShowLeaderboard(
	ctx *disgoplus.Ctx,
	container *di.Container,
	source leaderboardSourceType,
	page int,
	getItems LeaderboardDataFunc,
	formatter LeaderboardFormatFunc,
) {
	if source == LEADERBOARD_INTERACTION {
		disgoplus.Defer(ctx, true) //nolint:errcheck
	}

	items, total, err := getItems(ctx, page)
	if err != nil {
		Logger.Errorw("leaderboard: get items failed", "error", err, "guildID", ctx.GuildID)
		doError(ctx, source)

		return
	}

	if total == 0 {
		doTextResponse(ctx, source, "There is no leaderboard available yet for this server.")
		return
	}

	if len(items) == 0 {
		doTextResponse(ctx, source, fmt.Sprintf("No players found for page %d", page))
		return
	}

	doLeaderboardResponse(ctx, container, source, page, total, items, formatter)
}

func doLeaderboardResponse(
	ctx *disgoplus.Ctx,
	container *di.Container,
	source leaderboardSourceType,
	page int,
	total int,
	items []any,
	formatter LeaderboardFormatFunc,
) {
	embedColor := container.Get(static.DiEmbedColor).(int)
	cfg := container.Get(static.DiConfig).(*config.Config)
	bot := container.Get(static.DiBot).(*disgoplus.Bot)

	maxPage := int(math.Ceil(float64(total) / 10))

	footerParams := CreateEmbedFooterParams{IsVote: false}
	if maxPage > 1 {
		footerParams.Text = fmt.Sprintf("Page %d/%d", page, maxPage)
	}

	footer := CreateEmbedFooter(bot, &footerParams, cfg.OwnerID)

	guild, err := bot.Client().Rest.GetGuild(snowflake.MustParse(ctx.GuildID.String()), false)
	if err != nil || guild == nil {
		Logger.Errorw("leaderboard: get guild failed", "error", err, "guildID", ctx.GuildID)
		doError(ctx, source)

		return
	}

	description := ""
	for i, item := range items {
		description = fmt.Sprintf("%s\n%d. %s", description, i+1, formatter(ctx, item))
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
		buttons = append(buttons, discord.NewPrimaryButton("◀️", fmt.Sprintf("LEADERBOARD/%d", page-1)))
	}

	if page < maxPage {
		buttons = append(buttons, discord.NewPrimaryButton("▶️", fmt.Sprintf("LEADERBOARD/%d", page+1)))
	}

	var components []discord.LayoutComponent
	if len(buttons) > 0 {
		components = []discord.LayoutComponent{discord.NewActionRow(buttons...)}
	}

	if source == LEADERBOARD_MESSAGE_COMPONENT {
		emptyEmbeds := []discord.Embed{embed}
		if err := disgoplus.Update(ctx, discord.MessageUpdate{
			Embeds:     &emptyEmbeds,
			Components: &components,
		}); err != nil {
			Logger.Errorw("leaderboard: update response failed", "error", err, "guildID", ctx.GuildID)
		}

		return
	}

	if _, err := disgoplus.FollowUp(ctx, discord.MessageCreate{
		Embeds:     []discord.Embed{embed},
		Components: components,
		Flags:      discord.MessageFlagEphemeral,
	}); err != nil {
		Logger.Errorw("leaderboard: follow up response failed", "error", err, "guildID", ctx.GuildID)
	}
}

func doTextResponse(ctx *disgoplus.Ctx, source leaderboardSourceType, content string) {
	empty := []discord.Embed{}
	emptyC := []discord.LayoutComponent{}

	if source == LEADERBOARD_INTERACTION {
		disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
			Content:    content,
			Embeds:     empty,
			Components: emptyC,
			Flags:      discord.MessageFlagEphemeral,
		})

		return
	}

	disgoplus.Update(ctx, discord.MessageUpdate{ //nolint:errcheck
		Content:    &content,
		Embeds:     &empty,
		Components: &emptyC,
	})
}

func doError(ctx *disgoplus.Ctx, source leaderboardSourceType) {
	if source == LEADERBOARD_INTERACTION {
		disgoplus.InteractionError(ctx, true)
		return
	}

	if source == LEADERBOARD_MESSAGE_COMPONENT {
		disgoplus.MessageComponentError(ctx)
	}
}
