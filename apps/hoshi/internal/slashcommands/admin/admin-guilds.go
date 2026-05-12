package slashcommands

import (
	"fmt"
	"math"
	"strings"

	"github.com/FedorLap2006/disgolf"
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/utils"
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

func (m *AdminGuildsModule) list(ctx *disgolf.Ctx) {
	page := 1
	if opt, ok := ctx.Options["page"]; ok {
		page = int(opt.IntValue())
	}
	m.showList(ctx, page, false)
}

func (m *AdminGuildsModule) listPage(ctx *disgolf.Ctx) {
	page := 1
	if p, ok := ctx.MessageComponentOptions["page"]; ok {
		fmt.Sscanf(p, "%d", &page)
	}
	m.showList(ctx, page, true)
}

func (m *AdminGuildsModule) showList(ctx *disgolf.Ctx, page int, isComponent bool) {
	if !isComponent {
		utils.Defer(ctx, true)
	}

	guilds, total := m.guilds.GetData(page)

	if total == 0 {
		content := "There is no guild data available."
		if isComponent {
			utils.Update(ctx, &discordgo.InteractionResponseData{Content: content, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}})
		} else {
			utils.FollowUp(ctx, &discordgo.WebhookParams{Content: content}, true)
		}
		return
	}

	if len(guilds) == 0 {
		content := fmt.Sprintf("No guilds found for page %d", page)
		if isComponent {
			utils.Update(ctx, &discordgo.InteractionResponseData{Content: content, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}})
		} else {
			utils.FollowUp(ctx, &discordgo.WebhookParams{Content: content}, true)
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
		utils.Update(ctx, &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		})
	} else {
		utils.FollowUp(ctx, &discordgo.WebhookParams{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		}, true)
	}
}

func (m *AdminGuildsModule) Commands() []*disgolf.Command {
	return []*disgolf.Command{
		{
			Name:        "guilds",
			Description: "Get a list of guilds sorted by member count",
			Handler:     disgolf.HandlerFunc(m.list),
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

func (m *AdminGuildsModule) MessageComponents() []*disgolf.MessageComponent {
	return []*disgolf.MessageComponent{
		{
			CustomID: "ADMIN_GUILDS_LIST/:page",
			Handler:  disgolf.HandlerFunc(m.listPage),
		},
	}
}
