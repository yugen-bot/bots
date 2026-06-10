package leaderboard

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *LeaderboardModule) leaderboard(ctx *disgoplus.Ctx) {
	leaderboardType := "points"
	if v, ok := ctx.CommandData.OptString("type"); ok {
		leaderboardType = v
	}

	page := 1
	if v, ok := ctx.CommandData.OptInt("page"); ok {
		page = v
	}

	ephemeral := true
	if v, ok := ctx.CommandData.OptBool("ephemeral"); ok {
		ephemeral = v
	}

	disgoplus.Defer(ctx, ephemeral)
	m.showLeaderboard(ctx, leaderboardType, page, ephemeral, false)
}

func (m *LeaderboardModule) leaderboardPage(ctx *disgoplus.Ctx) {
	leaderboardType := "points"
	if t := ctx.MessageComponentOptions["type"]; t != "" {
		leaderboardType = t
	}

	page := 1
	if p, err := strconv.Atoi(
		ctx.MessageComponentOptions["page"],
	); err == nil &&
		p > 0 {
		page = p
	}

	m.showLeaderboard(ctx, leaderboardType, page, true, true)
}

func (m *LeaderboardModule) showLeaderboard(
	ctx *disgoplus.Ctx,
	leaderboardType string,
	page int,
	ephemeral bool,
	isComponent bool,
) {
	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.GuildID.String(),
	)
	if err != nil || settings == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if settings.ChannelID == nil || *settings.ChannelID == "" {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	players, total, err := m.points.GetLeaderboard(
		context.Background(),
		ctx.GuildID.String(),
		leaderboardType,
		page,
	)
	if err != nil {
		disgoplus.InteractionError(ctx, ephemeral)
		return
	}

	maxPage := int(math.Ceil(float64(total) / 10))
	if maxPage == 0 {
		maxPage = 1
	}

	typeLabel := strings.ToUpper(leaderboardType[:1]) + leaderboardType[1:]

	var sb strings.Builder

	for i, p := range players {
		rank := (page-1)*10 + i + 1

		var value int

		switch leaderboardType {
		case "wins":
			value = p.Wins
		case "participated":
			value = p.Participated
		default:
			value = p.Points
		}

		fmt.Fprintf(&sb, "%d. <@%s>: **%d**\n", rank, p.UserID, value)
	}

	if sb.Len() == 0 {
		sb.WriteString("No players found.")
	}

	bot := m.container.Get(sharedStatic.DiBot).(*disgoplus.Bot)

	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		"",
	)
	if footer != nil && maxPage > 1 {
		footer.Text = fmt.Sprintf("Page %d/%d | %s", page, maxPage, footer.Text)
	}

	embed := discord.NewEmbed().
		WithTitle(fmt.Sprintf("Koto Leaderboard — %s", typeLabel)).
		WithColor(localStatic.EmbedColor).
		WithDescription(sb.String()).
		WithEmbedFooter(footer)

	if guild, ok := ctx.Client.Caches.Guild(ctx.GuildID); ok {
		if iconURL := guild.IconURL(); iconURL != nil {
			embed = embed.WithThumbnail(*iconURL)
		}
	}

	var buttons []discord.InteractiveComponent
	if page > 1 {
		buttons = append(
			buttons,
			discord.NewPrimaryButton(
				"◀️",
				fmt.Sprintf("LEADERBOARD/%s/%d", leaderboardType, page-1),
			),
		)
	}

	if page < maxPage {
		buttons = append(
			buttons,
			discord.NewPrimaryButton(
				"▶️",
				fmt.Sprintf("LEADERBOARD/%s/%d", leaderboardType, page+1),
			),
		)
	}

	components := []discord.LayoutComponent{}
	if len(buttons) > 0 {
		components = append(components, discord.NewActionRow(buttons...))
	}

	if isComponent {
		embeds := []discord.Embed{embed}
		disgoplus.Update(ctx, discord.MessageUpdate{
			Embeds:     &embeds,
			Components: &components,
		})
	} else {
		flags := discord.MessageFlags(0)
		if ephemeral {
			flags = discord.MessageFlagEphemeral
		}
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Embeds:     []discord.Embed{embed},
			Components: components,
			Flags:      flags,
		})
	}
}
