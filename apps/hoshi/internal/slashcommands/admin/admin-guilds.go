package slashcommands

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
)

type AdminGuildsModule struct {
	container *di.Container
	guilds    *services.GuildsService
}

func GetAdminGuildsModule(container *di.Container) *AdminGuildsModule {
	return &AdminGuildsModule{
		container: container,
		guilds:    container.Get(localStatic.DiGuilds).(*services.GuildsService),
	}
}

func (m *AdminGuildsModule) list(ctx *discordgoplus.Ctx) {
	page := 1
	if opt, ok := ctx.Options["page"]; ok {
		page = int(opt.IntValue())
	}
	m.showList(ctx, page, false)
}

func (m *AdminGuildsModule) listPage(ctx *discordgoplus.Ctx) {
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

func (m *AdminGuildsModule) showList(ctx *discordgoplus.Ctx, page int, isComponent bool) {
	if !isComponent {
		discordgoplus.Defer(ctx, true)
	}

	guilds, total := m.guilds.GetData(page)

	if total == 0 {
		content := "There is no guild data available."
		if isComponent {
			discordgoplus.Update(ctx, &discordgo.InteractionResponseData{Content: content, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}})
		} else {
			discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{Content: content}, true)
		}
		return
	}

	if len(guilds) == 0 {
		content := fmt.Sprintf("No guilds found for page %d", page)
		if isComponent {
			discordgoplus.Update(ctx, &discordgo.InteractionResponseData{Content: content, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}})
		} else {
			discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{Content: content}, true)
		}
		return
	}

	maxPage := int(math.Ceil(float64(total) / 10))

	lines := make([]string, len(guilds))
	for i, g := range guilds {
		lines[i] = fmt.Sprintf("%d. %s: **%d**", (page-1)*10+(i+1), g.Name, g.MemberCount)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Hoshi guilds",
		Description: strings.Join(lines, "\n"),
	}

	if maxPage > 1 {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d/%d", page, maxPage),
		}
	}

	var buttons []discordgo.MessageComponent
	if page > 1 {
		buttons = append(buttons, discordgo.Button{
			CustomID: fmt.Sprintf("ADMIN_GUILDS_LIST/%d", page-1),
			Style:    discordgo.PrimaryButton,
			Label:    "◀️",
		})
	}
	if page < maxPage {
		buttons = append(buttons, discordgo.Button{
			CustomID: fmt.Sprintf("ADMIN_GUILDS_LIST/%d", page+1),
			Style:    discordgo.PrimaryButton,
			Label:    "▶️",
		})
	}

	components := []discordgo.MessageComponent{}
	if len(buttons) > 0 {
		components = append(components, discordgo.ActionsRow{Components: buttons})
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

func (m *AdminGuildsModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "guilds",
			Description: "Get a list of guilds sorted by member count",
			Handler:     discordgoplus.HandlerFunc(m.list),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "page",
					Description: "View a specific page.",
					Required:    false,
				},
			},
		},
	}
}

func (m *AdminGuildsModule) MessageComponents() []*discordgoplus.MessageComponent {
	return []*discordgoplus.MessageComponent{
		{
			CustomID: "ADMIN_GUILDS_LIST/:page",
			Handler:  discordgoplus.HandlerFunc(m.listPage),
		},
	}
}
