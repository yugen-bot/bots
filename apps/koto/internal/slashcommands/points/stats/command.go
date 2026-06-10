package stats

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func boolPtr(b bool) *bool { return &b }

func (m *StatsModule) points(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	settings, err := m.settingsSvc.GetByGuildID(
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

	player, err := m.pointsSvc.GetPlayer(
		context.Background(),
		ctx.GuildID.String(),
		ctx.Member.User.ID.String(),
		false,
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	playerHints, _ := m.hintsSvc.GetPlayerHintsByUserID(
		context.Background(),
		ctx.Member.User.ID.String(),
	)

	bot := m.container.Get(sharedStatic.DiBot).(*disgoplus.Bot)
	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		"",
	)

	hintsValue := "0/3"
	if playerHints != nil {
		hintsValue = fmt.Sprintf(
			"%s/%s",
			strconv.FormatFloat(playerHints.Hints, 'f', -1, 64),
			strconv.FormatFloat(playerHints.MaxHints, 'f', -1, 64),
		)
	}

	embed := discord.NewEmbed().
		WithTitle("Your Koto Stats").
		WithColor(localStatic.EmbedColor).
		WithFields(
			discord.EmbedField{Name: "Points", Value: fmt.Sprintf("%d", player.Points), Inline: boolPtr(true)},
			discord.EmbedField{Name: "Games Participated", Value: fmt.Sprintf("%d", player.Participated), Inline: boolPtr(true)},
			discord.EmbedField{Name: "Games Won", Value: fmt.Sprintf("%d", player.Wins), Inline: boolPtr(true)},
			discord.EmbedField{Name: "Hints", Value: hintsValue, Inline: boolPtr(true)},
		).
		WithEmbedFooter(footer)

	disgoplus.FollowUp(
		ctx,
		discord.MessageCreate{
			Embeds: []discord.Embed{embed},
			Flags:  discord.MessageFlagEphemeral,
		},
	)
}
