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
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	localUtils "jurien.dev/yugen/hoshi/internal/utils"
	"jurien.dev/yugen/shared/static"
)

type StarboardListModule struct {
	container *di.Container
	starboard *services.StarboardService
}

func GetStarboardListModule(container *di.Container) *StarboardListModule {
	return &StarboardListModule{
		container: container,
		starboard: container.Get(localStatic.DiStarboard).(*services.StarboardService),
	}
}

func (m *StarboardListModule) list(ctx *discordgoplus.Ctx) {
	page := 1
	if opt, ok := ctx.Options["page"]; ok {
		page = int(opt.IntValue())
	}

	m.showList(ctx, page, false)
}

func (m *StarboardListModule) listPage(ctx *discordgoplus.Ctx) {
	page := 1

	if p, ok := ctx.MessageComponentOptions["page"]; ok {
		page64, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			page = 1 // fallback to first page
		} else {
			page = int(page64)
		}
	}

	m.showList(ctx, page, true)
}

func (m *StarboardListModule) showList(
	ctx *discordgoplus.Ctx,
	page int,
	isComponent bool,
) {
	if !isComponent {
		discordgoplus.Defer(ctx, true)
	}

	bot := m.container.Get(static.DiBot).(*discordgoplus.Bot)

	items, total, err := m.starboard.GetStarboards(
		context.Background(),
		ctx.Interaction.GuildID,
		page,
	)
	if err != nil {
		if isComponent {
			discordgoplus.MessageComponentError(ctx)
		} else {
			discordgoplus.InteractionError(ctx, true)
		}

		return
	}

	if total == 0 {
		content := "No starboards have been configured yet."
		if isComponent {
			discordgoplus.Update(
				ctx,
				&discordgo.InteractionResponseData{
					Content:    content,
					Embeds:     []*discordgo.MessageEmbed{},
					Components: []discordgo.MessageComponent{},
				},
			)
		} else {
			discordgoplus.FollowUp(
				ctx,
				&discordgo.WebhookParams{Content: content},
				true,
			)
		}

		return
	}

	if len(items) == 0 {
		content := fmt.Sprintf("No starboards found for page %d", page)
		if isComponent {
			discordgoplus.Update(
				ctx,
				&discordgo.InteractionResponseData{
					Content:    content,
					Embeds:     []*discordgo.MessageEmbed{},
					Components: []discordgo.MessageComponent{},
				},
			)
		} else {
			discordgoplus.FollowUp(
				ctx,
				&discordgo.WebhookParams{Content: content},
				true,
			)
		}

		return
	}

	maxPage := int(math.Ceil(float64(total) / 10))

	// Build display rows
	ids := make([]string, len(items))
	emojiSources := make([]string, len(items))
	destinations := make([]string, len(items))

	for i, c := range items {
		ids[i] = fmt.Sprintf("%d", c.ID)

		_, _, display, unicode := localUtils.ResolveEmoji(c.SourceEmoji, bot)

		emojiDisplay := c.SourceEmoji
		if !unicode {
			emojiDisplay = display
		}

		src := "Anywhere"
		if sid, ok := c.SourceChannelID(); ok {
			src = fmt.Sprintf("<#%s>", sid)
		}

		emojiSources[i] = fmt.Sprintf("%s | %s", emojiDisplay, src)
		destinations[i] = fmt.Sprintf("<#%s>", c.TargetChannelID)
	}

	embed := &discordgo.MessageEmbed{
		Color: localStatic.EmbedColor,
		Title: fmt.Sprintf("Starboards for %s", ctx.Interaction.GuildID),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: strings.Join(ids, "\n"), Inline: true},
			{
				Name:   "Emoji | Source",
				Value:  strings.Join(emojiSources, "\n"),
				Inline: true,
			},
			{
				Name:   "Destination",
				Value:  strings.Join(destinations, "\n"),
				Inline: true,
			},
		},
	}

	if maxPage > 1 {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d/%d", page, maxPage),
		}
	}

	var buttons []discordgo.MessageComponent
	if page > 1 {
		buttons = append(buttons, discordgo.Button{
			CustomID: fmt.Sprintf("STARBOARD_LIST/%d", page-1),
			Style:    discordgo.PrimaryButton,
			Label:    "◀️",
		})
	}

	if page < maxPage {
		buttons = append(buttons, discordgo.Button{
			CustomID: fmt.Sprintf("STARBOARD_LIST/%d", page+1),
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
		}, true)
	}
}

func (m *StarboardListModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "list",
			Description: "List the starboards",
			Handler:     discordgoplus.HandlerFunc(m.list),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "page",
					Description: "The page to view",
					Required:    false,
					MinValue:    func() *float64 { v := float64(1); return &v }(),
				},
			},
		},
	}
}

func (m *StarboardListModule) MessageComponents() []*discordgoplus.MessageComponent {
	return []*discordgoplus.MessageComponent{
		{
			CustomID: "STARBOARD_LIST/:page",
			Handler:  discordgoplus.HandlerFunc(m.listPage),
		},
	}
}
