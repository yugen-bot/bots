package slashcommands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type VoteModule struct {
	container *di.Container
}

func GetVoteModule(container *di.Container) *VoteModule {
	return &VoteModule{container: container}
}

func (m *VoteModule) Run(ctx *disgoplus.Ctx) {
	if err := disgoplus.Defer(ctx); err != nil {
		utils.Logger.Error(err)
		return
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiClient).(*disgoplus.Bot)

	var botName string
	if self, ok := bot.Client().Caches.SelfUser(); ok {
		botName = self.Username
	}

	var voteReward string
	voteRewardFunc := m.container.Get(static.DiVoteReward).(func(userId string) string)
	if voteRewardFunc != nil && ctx.Member != nil {
		voteReward = voteRewardFunc(ctx.Member.User.ID.String())
	}
	if len(voteReward) > 0 {
		voteReward = "\n" + voteReward
	}

	footer := utils.CreateEmbedFooter(bot, &utils.CreateEmbedFooterParams{IsVote: true}, cfg.OwnerID)

	embedColor := m.container.Get(static.DiEmbedColor).(int)

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle("Vote information").
		WithDescription(fmt.Sprintf(
			`Like what %s is doing and want to support it's growth?
Please use any of the links below to vote for %s!%s`,
			botName, botName, voteReward,
		)).
		WithEmbedFooter(footer)

	var buttons []discord.InteractiveComponent
	if cfg.TopGGVoteLink != "" {
		buttons = append(buttons, discord.NewLinkButton("Vote on Top.GG", cfg.TopGGVoteLink))
	}
	if cfg.DiscordBotListVoteLink != "" {
		buttons = append(buttons, discord.NewLinkButton("Vote on Discord Bot List", cfg.DiscordBotListVoteLink))
	}

	msg := discord.NewMessageCreate().AddEmbeds(embed)
	if len(buttons) > 0 {
		msg = msg.AddActionRow(buttons...)
	}

	if _, err := disgoplus.FollowUp(ctx, msg); err != nil {
		utils.Logger.Error(err)
	}
}

func (m *VoteModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "vote",
			Description: "Vote for the bot!",
			Handler:     disgoplus.HandlerFunc(m.Run),
		},
	}
}
