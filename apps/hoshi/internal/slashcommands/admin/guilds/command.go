package guilds

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (m *GuildsModule) list(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	page := 1
	if v, ok := data.OptInt("page"); ok {
		page = v
	}

	return m.showList(e, nil, page)
}

func (m *GuildsModule) listPage(
	_ discord.ButtonInteractionData,
	e *handler.ComponentEvent,
) error {
	page := 1

	if p, ok := e.Vars["page"]; ok {
		page64, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			page = 1
		} else {
			page = int(page64)
		}
	}

	return m.showListComponent(e, page)
}

func (m *GuildsModule) showList(
	e *handler.CommandEvent,
	_ interface{},
	page int,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	guilds, total := m.guilds.GetData(page)

	if total == 0 {
		_, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "There is no guild data available.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	if len(guilds) == 0 {
		_, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf("No guilds found for page %d", page),
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	maxPage := int(math.Ceil(float64(total) / 10))

	lines := make([]string, len(guilds))
	for i, g := range guilds {
		lines[i] = fmt.Sprintf(
			"%d. %s: **%d**",
			(page-1)*10+(i+1),
			g.Name,
			g.MemberCount,
		)
	}

	embed := discord.NewEmbed().
		WithTitle("Hoshi guilds").
		WithDescription(strings.Join(lines, "\n"))

	if maxPage > 1 {
		embed = embed.WithEmbedFooter(&discord.EmbedFooter{
			Text: fmt.Sprintf("Page %d/%d", page, maxPage),
		})
	}

	var buttons []discord.InteractiveComponent
	if page > 1 {
		buttons = append(
			buttons,
			discord.NewPrimaryButton(
				"◀️",
				fmt.Sprintf("/ADMIN_GUILDS_LIST/%d", page-1),
			),
		)
	}

	if page < maxPage {
		buttons = append(
			buttons,
			discord.NewPrimaryButton(
				"▶️",
				fmt.Sprintf("/ADMIN_GUILDS_LIST/%d", page+1),
			),
		)
	}

	components := []discord.LayoutComponent{}
	if len(buttons) > 0 {
		components = append(components, discord.NewActionRow(buttons...))
	}

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Embeds:     []discord.Embed{embed},
		Components: components,
		Flags:      discord.MessageFlagEphemeral,
	})

	return err
}

func (m *GuildsModule) showListComponent(
	e *handler.ComponentEvent,
	page int,
) error {
	guilds, total := m.guilds.GetData(page)

	if total == 0 {
		content := "There is no guild data available."
		empty := []discord.Embed{}
		emptyComponents := []discord.LayoutComponent{}

		return e.UpdateMessage(discord.MessageUpdate{
			Content:    &content,
			Embeds:     &empty,
			Components: &emptyComponents,
		})
	}

	if len(guilds) == 0 {
		content := fmt.Sprintf("No guilds found for page %d", page)
		empty := []discord.Embed{}
		emptyComponents := []discord.LayoutComponent{}

		return e.UpdateMessage(discord.MessageUpdate{
			Content:    &content,
			Embeds:     &empty,
			Components: &emptyComponents,
		})
	}

	maxPage := int(math.Ceil(float64(total) / 10))

	lines := make([]string, len(guilds))
	for i, g := range guilds {
		lines[i] = fmt.Sprintf(
			"%d. %s: **%d**",
			(page-1)*10+(i+1),
			g.Name,
			g.MemberCount,
		)
	}

	embed := discord.NewEmbed().
		WithTitle("Hoshi guilds").
		WithDescription(strings.Join(lines, "\n"))

	if maxPage > 1 {
		embed = embed.WithEmbedFooter(&discord.EmbedFooter{
			Text: fmt.Sprintf("Page %d/%d", page, maxPage),
		})
	}

	var buttons []discord.InteractiveComponent
	if page > 1 {
		buttons = append(
			buttons,
			discord.NewPrimaryButton(
				"◀️",
				fmt.Sprintf("/ADMIN_GUILDS_LIST/%d", page-1),
			),
		)
	}

	if page < maxPage {
		buttons = append(
			buttons,
			discord.NewPrimaryButton(
				"▶️",
				fmt.Sprintf("/ADMIN_GUILDS_LIST/%d", page+1),
			),
		)
	}

	components := []discord.LayoutComponent{}
	if len(buttons) > 0 {
		components = append(components, discord.NewActionRow(buttons...))
	}

	embeds := []discord.Embed{embed}

	return e.UpdateMessage(discord.MessageUpdate{
		Embeds:     &embeds,
		Components: &components,
	})
}
