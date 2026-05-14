package slashcommands

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type LeaderboardModule struct {
	container *di.Container
	points    *services.PointsService
	settings  *services.SettingsService
}

func GetLeaderboardModule(container *di.Container) *LeaderboardModule {
	return &LeaderboardModule{
		container: container,
		points:    container.Get(localStatic.DiPoints).(*services.PointsService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *LeaderboardModule) leaderboard(ctx *discordgoplus.Ctx) {
	leaderboardType := "points"
	if opt, ok := ctx.Options["type"]; ok {
		leaderboardType = opt.StringValue()
	}

	page := 1
	if opt, ok := ctx.Options["page"]; ok {
		page = int(opt.IntValue())
	}

	ephemeral := true
	if opt, ok := ctx.Options["ephemeral"]; ok {
		ephemeral = opt.BoolValue()
	}

	discordgoplus.Defer(ctx, ephemeral)
	m.showLeaderboard(ctx, leaderboardType, page, ephemeral, false)
}

func (m *LeaderboardModule) leaderboardPage(ctx *discordgoplus.Ctx) {
	// CustomID format: LEADERBOARD/:data where data = "type/page"
	data := ctx.MessageComponentOptions["data"]
	leaderboardType := "points"
	page := 1

	parts := strings.SplitN(data, "/", 2)
	if len(parts) >= 1 && parts[0] != "" {
		leaderboardType = parts[0]
	}
	if len(parts) >= 2 {
		if p, err := strconv.Atoi(parts[1]); err == nil && p > 0 {
			page = p
		}
	}

	m.showLeaderboard(ctx, leaderboardType, page, true, true)
}

func (m *LeaderboardModule) showLeaderboard(
	ctx *discordgoplus.Ctx,
	leaderboardType string,
	page int,
	ephemeral bool,
	isComponent bool,
) {
	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil || settings == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	channelID, ok := settings.ChannelID()
	if !ok || channelID == "" {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	players, total, err := m.points.GetLeaderboard(
		context.Background(),
		ctx.Interaction.GuildID,
		leaderboardType,
		page,
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, ephemeral)
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

	bot := m.container.Get(sharedStatic.DiBot).(*discordgoplus.Bot)
	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		"",
	)
	if footer != nil && maxPage > 1 {
		footer.Text = fmt.Sprintf("Page %d/%d | %s", page, maxPage, footer.Text)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Koto Leaderboard — %s", typeLabel),
		Color:       localStatic.EmbedColor,
		Description: sb.String(),
		Footer:      footer,
	}

	var buttons []discordgo.MessageComponent
	if page > 1 {
		buttons = append(buttons, discordgo.Button{
			CustomID: fmt.Sprintf("LEADERBOARD/%s/%d", leaderboardType, page-1),
			Style:    discordgo.PrimaryButton,
			Label:    "◀️",
		})
	}
	if page < maxPage {
		buttons = append(buttons, discordgo.Button{
			CustomID: fmt.Sprintf("LEADERBOARD/%s/%d", leaderboardType, page+1),
			Style:    discordgo.PrimaryButton,
			Label:    "▶️",
		})
	}

	components := []discordgo.MessageComponent{}
	if len(buttons) > 0 {
		components = append(
			components,
			discordgo.ActionsRow{Components: buttons},
		)
	}

	if isComponent {
		discordgoplus.Update(ctx, &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		})
	} else {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		}, ephemeral)
	}
}

func (m *LeaderboardModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "leaderboard",
			Description: "View the Koto leaderboard",
			Handler:     discordgoplus.HandlerFunc(m.leaderboard),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "Leaderboard type",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Points", Value: "points"},
						{Name: "Wins", Value: "wins"},
						{Name: "Participated", Value: "participated"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "page",
					Description: "Page number",
					Required:    false,
					MinValue:    func() *float64 { v := float64(1); return &v }(),
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "ephemeral",
					Description: "Show only to you",
					Required:    false,
				},
			},
		},
	}
}

func (m *LeaderboardModule) MessageComponents() []*discordgoplus.MessageComponent {
	return []*discordgoplus.MessageComponent{
		{
			// CustomID pattern: LEADERBOARD/:data where data encodes "type/page"
			CustomID: "LEADERBOARD/:data",
			Handler:  discordgoplus.HandlerFunc(m.leaderboardPage),
		},
	}
}
