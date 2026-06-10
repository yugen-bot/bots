package leaderboard

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/koto/internal/ent"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *LeaderboardModule) leaderboard(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	leaderboardType := "points"
	if v, ok := data.OptString("type"); ok {
		leaderboardType = v
	}

	page := 1
	if v, ok := data.OptInt("page"); ok {
		page = v
	}

	ephemeral := true
	if v, ok := data.OptBool("ephemeral"); ok {
		ephemeral = v
	}

	if err := e.DeferCreateMessage(ephemeral); err != nil {
		return fmt.Errorf("leaderboard: defer: %w", err)
	}

	return m.showLeaderboardCommand(e, leaderboardType, page, ephemeral)
}

func (m *LeaderboardModule) leaderboardPage(e *handler.ComponentEvent) error {
	leaderboardType := "points"
	if t := e.Vars["type"]; t != "" {
		leaderboardType = t
	}

	page := 1
	if p, err := strconv.Atoi(e.Vars["page"]); err == nil && p > 0 {
		page = p
	}

	return m.showLeaderboardComponent(e, leaderboardType, page)
}

func (m *LeaderboardModule) buildLeaderboardContent(
	players []*ent.PlayerStats,
	leaderboardType string,
	page int,
	total int,
) (discord.Embed, []discord.LayoutComponent) {
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

	var buttons []discord.InteractiveComponent
	if page > 1 {
		buttons = append(
			buttons,
			discord.NewPrimaryButton(
				"◀️",
				fmt.Sprintf(customIDLeaderboardPage, leaderboardType, page-1),
			),
		)
	}

	if page < maxPage {
		buttons = append(
			buttons,
			discord.NewPrimaryButton(
				"▶️",
				fmt.Sprintf(customIDLeaderboardPage, leaderboardType, page+1),
			),
		)
	}

	components := []discord.LayoutComponent{}
	if len(buttons) > 0 {
		components = append(components, discord.NewActionRow(buttons...))
	}

	return embed, components
}

func (m *LeaderboardModule) showLeaderboardCommand(
	e *handler.CommandEvent,
	leaderboardType string,
	page int,
	ephemeral bool,
) error {
	settings, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil || settings == nil {
		return localUtils.ReplyNoSettings(e, true)
	}

	if settings.ChannelID == nil || *settings.ChannelID == "" {
		return localUtils.ReplyNoSettings(e, true)
	}

	players, total, err := m.points.GetLeaderboard(
		context.Background(),
		e.GuildID().String(),
		leaderboardType,
		page,
	)
	if err != nil {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("leaderboard: send followup: %w", sendErr)
		}

		return nil
	}

	embed, components := m.buildLeaderboardContent(
		players, leaderboardType, page, total,
	)

	if guild, ok := e.Client().Caches.Guild(*e.GuildID()); ok {
		if iconURL := guild.IconURL(); iconURL != nil {
			embed = embed.WithThumbnail(*iconURL)
		}
	}

	flags := discord.MessageFlags(0)
	if ephemeral {
		flags = discord.MessageFlagEphemeral
	}

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Embeds:     []discord.Embed{embed},
		Components: components,
		Flags:      flags,
	})
	if sendErr != nil {
		return fmt.Errorf("leaderboard: send followup: %w", sendErr)
	}

	return nil
}

func (m *LeaderboardModule) showLeaderboardComponent(
	e *handler.ComponentEvent,
	leaderboardType string,
	page int,
) error {
	settings, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil || settings == nil {
		if createErr := e.CreateMessage(discord.MessageCreate{
			Content: localUtils.NoSettingsDescription,
			Flags:   discord.MessageFlagEphemeral,
		}); createErr != nil {
			return fmt.Errorf(
				"leaderboard component: create message: %w",
				createErr,
			)
		}

		return nil
	}

	if settings.ChannelID == nil || *settings.ChannelID == "" {
		if createErr := e.CreateMessage(discord.MessageCreate{
			Content: localUtils.NoSettingsDescription,
			Flags:   discord.MessageFlagEphemeral,
		}); createErr != nil {
			return fmt.Errorf(
				"leaderboard component: create message: %w",
				createErr,
			)
		}

		return nil
	}

	players, total, err := m.points.GetLeaderboard(
		context.Background(),
		e.GuildID().String(),
		leaderboardType,
		page,
	)
	if err != nil {
		if createErr := e.CreateMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); createErr != nil {
			return fmt.Errorf(
				"leaderboard component: create message: %w",
				createErr,
			)
		}

		return nil
	}

	embed, components := m.buildLeaderboardContent(
		players, leaderboardType, page, total,
	)

	if guild, ok := e.Client().Caches.Guild(*e.GuildID()); ok {
		if iconURL := guild.IconURL(); iconURL != nil {
			embed = embed.WithThumbnail(*iconURL)
		}
	}

	embeds := []discord.Embed{embed}

	if updateErr := e.UpdateMessage(discord.MessageUpdate{
		Embeds:     &embeds,
		Components: &components,
	}); updateErr != nil {
		return fmt.Errorf(
			"leaderboard component: update message: %w",
			updateErr,
		)
	}

	return nil
}
