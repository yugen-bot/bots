package slashcommands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type VoteModule struct {
	container *di.Container
}

func GetVoteModule(container *di.Container) *VoteModule {
	return &VoteModule{
		container: container,
	}
}

func (m *VoteModule) Run(ctx *discordgoplus.Ctx) {
	err := discordgoplus.Defer(ctx)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	user := ctx.Interaction.Member.User
	bot := ctx.State.User
	name := bot.Username

	voteReward := ""

	voteRewardFunc := m.container.Get(static.DiVoteReward).(func(userId string) string)
	embedColor := m.container.Get(static.DiEmbedColor).(int)

	if voteRewardFunc != nil && user != nil {
		voteReward = voteRewardFunc(user.ID)
	}

	if len(voteReward) > 0 {
		voteReward = "\n" + voteReward
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)

	footer := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*discordgoplus.Bot),
		&utils.CreateEmbedFooterParams{
			IsVote: true,
		},
		cfg.OwnerID,
	)

	embed := &discordgo.MessageEmbed{
		Color: embedColor,
		Title: "Vote information",
		Description: fmt.Sprintf(`Like what %s is doing and want to support it's growth?
Please use any of the links below to vote for %s!%s`, name, name, voteReward),
		Footer: footer,
	}

	components := []discordgo.MessageComponent{}

	topGGVoteLink := cfg.TopGGVoteLink
	discordBotListVoteLink := cfg.DiscordBotListVoteLink

	if len(topGGVoteLink) > 0 {
		components = append(components, discordgo.Button{
			Style: discordgo.LinkButton,
			Label: "Vote on Top.GG",
			URL:   topGGVoteLink,
		})
	}

	if len(discordBotListVoteLink) > 0 {
		components = append(components, discordgo.Button{
			Style: discordgo.LinkButton,
			Label: "Vote on Discord Bot List",
			URL:   discordBotListVoteLink,
		})
	}

	messageComponents := []discordgo.MessageComponent{}

	if len(components) > 0 {
		messageComponents = append(messageComponents, discordgo.ActionsRow{
			Components: components,
		})
	}

	err = discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: messageComponents,
	})
	if err != nil {
		utils.Logger.Error(err)
	}
}

func (m *VoteModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "vote",
			Description: "Vote for the bot!",
			Handler:     discordgoplus.HandlerFunc(m.Run),
		},
	}
}
