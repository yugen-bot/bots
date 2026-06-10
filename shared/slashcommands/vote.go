package slashcommands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *VoteModule) run(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(false); err != nil {
		return fmt.Errorf("defer create message: %w", err)
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)

	var botName string
	if self, ok := bot.Client().Caches.SelfUser(); ok {
		botName = self.Username
	}

	var voteReward string

	voteRewardFunc := m.container.Get(static.DiVoteReward).(func(userId string) string)
	if voteRewardFunc != nil && e.Member() != nil {
		voteReward = voteRewardFunc(e.Member().User.ID.String())
	}

	if len(voteReward) > 0 {
		voteReward = "\n" + voteReward
	}

	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{IsVote: true},
		cfg.OwnerID,
	)

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
		buttons = append(
			buttons,
			discord.NewLinkButton("Vote on Top.GG", cfg.TopGGVoteLink),
		)
	}

	if cfg.DiscordBotListVoteLink != "" {
		buttons = append(
			buttons,
			discord.NewLinkButton(
				"Vote on Discord Bot List",
				cfg.DiscordBotListVoteLink,
			),
		)
	}

	msg := discord.NewMessageCreate().AddEmbeds(embed)
	if len(buttons) > 0 {
		msg = msg.AddActionRow(buttons...)
	}

	if _, err := e.CreateFollowupMessage(msg); err != nil {
		return fmt.Errorf("create followup message: %w", err)
	}

	return nil
}

func (m *VoteModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "vote",
			Description: "Vote for the bot!",
		},
	}
}

func (m *VoteModule) Register(r handler.Router) {
	r.SlashCommand("/vote", m.run)
}
