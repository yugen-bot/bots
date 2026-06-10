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
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("guilds list: defer: %w", err)
	}

	page := 1
	if v, ok := data.OptInt("page"); ok {
		page = v
	}

	return m.showList(e, nil, page, false)
}

func (m *GuildsModule) listPage(e *handler.ComponentEvent) error {
	page := 1

	if p, err := strconv.Atoi(e.Vars["page"]); err == nil {
		page = p
	}

	return m.showListComponent(e, page)
}

func (m *GuildsModule) showList(
	e *handler.CommandEvent,
	_ any,
	page int,
	isComponent bool,
) error {
	_ = isComponent // always false for CommandEvent

	guilds, total := m.guilds.GetData(page)

	if total == 0 {
		content := "There is no guild data available."

		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: content,
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("guilds list: send followup: %w", sendErr)
		}

		return nil
	}

	if len(guilds) == 0 {
		content := fmt.Sprintf("No guilds found for page %d", page)

		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: content,
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("guilds list: send followup: %w", sendErr)
		}

		return nil
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
		WithTitle("Koto guilds").
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

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Embeds:     []discord.Embed{embed},
		Components: components,
		Flags:      discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("guilds list: send followup: %w", sendErr)
	}

	return nil
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

		if updateErr := e.UpdateMessage(discord.MessageUpdate{
			Content:    &content,
			Embeds:     &empty,
			Components: &emptyComponents,
		}); updateErr != nil {
			return fmt.Errorf("guilds list: update message: %w", updateErr)
		}

		return nil
	}

	if len(guilds) == 0 {
		content := fmt.Sprintf("No guilds found for page %d", page)
		empty := []discord.Embed{}
		emptyComponents := []discord.LayoutComponent{}

		if updateErr := e.UpdateMessage(discord.MessageUpdate{
			Content:    &content,
			Embeds:     &empty,
			Components: &emptyComponents,
		}); updateErr != nil {
			return fmt.Errorf("guilds list: update message: %w", updateErr)
		}

		return nil
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
		WithTitle("Koto guilds").
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

	if updateErr := e.UpdateMessage(discord.MessageUpdate{
		Embeds:     &embeds,
		Components: &components,
	}); updateErr != nil {
		return fmt.Errorf("guilds list: update message: %w", updateErr)
	}

	return nil
}
