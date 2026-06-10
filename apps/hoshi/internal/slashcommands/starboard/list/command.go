package list

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/hoshi/internal/ent"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	localUtils "jurien.dev/yugen/hoshi/internal/utils"
	"jurien.dev/yugen/shared/static"
)

func (m *ListModule) list(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	page := 1
	if v, ok := data.OptInt("page"); ok {
		page = v
	}

	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)
	guildID := (*e.GuildID()).String()

	items, total, err := m.starboard.GetStarboards(context.Background(), guildID, page)
	if err != nil {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if ferr != nil {
			return ferr
		}
		return err
	}

	if total == 0 {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "No starboards have been configured yet.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return ferr
	}

	if len(items) == 0 {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf("No starboards found for page %d", page),
			Flags:   discord.MessageFlagEphemeral,
		})
		return ferr
	}

	embed, components := buildListContent(bot, items, total, page, guildID)

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Embeds:     []discord.Embed{embed},
		Components: components,
		Flags:      discord.MessageFlagEphemeral,
	})
	return err
}

func (m *ListModule) listPage(_ discord.ButtonInteractionData, e *handler.ComponentEvent) error {
	page := 1

	if p, ok := e.Vars["page"]; ok {
		page64, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			page = 1
		} else {
			page = int(page64)
		}
	}

	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)
	guildID := (*e.GuildID()).String()

	items, total, err := m.starboard.GetStarboards(context.Background(), guildID, page)
	if err != nil {
		content := "Something went wrong."
		empty := []discord.Embed{}
		emptyComponents := []discord.LayoutComponent{}
		return e.UpdateMessage(discord.MessageUpdate{
			Content:    &content,
			Embeds:     &empty,
			Components: &emptyComponents,
		})
	}

	if total == 0 {
		content := "No starboards have been configured yet."
		empty := []discord.Embed{}
		emptyComponents := []discord.LayoutComponent{}
		return e.UpdateMessage(discord.MessageUpdate{
			Content:    &content,
			Embeds:     &empty,
			Components: &emptyComponents,
		})
	}

	if len(items) == 0 {
		content := fmt.Sprintf("No starboards found for page %d", page)
		empty := []discord.Embed{}
		emptyComponents := []discord.LayoutComponent{}
		return e.UpdateMessage(discord.MessageUpdate{
			Content:    &content,
			Embeds:     &empty,
			Components: &emptyComponents,
		})
	}

	embed, components := buildListContent(bot, items, total, page, guildID)

	embeds := []discord.Embed{embed}
	return e.UpdateMessage(discord.MessageUpdate{
		Embeds:     &embeds,
		Components: &components,
	})
}

func buildListContent(
	bot *disgoplus.Bot,
	items []*ent.Starboards,
	total, page int,
	guildID string,
) (discord.Embed, []discord.LayoutComponent) {
	maxPage := int(math.Ceil(float64(total) / 10))

	ids := make([]string, len(items))
	emojiSources := make([]string, len(items))
	destinations := make([]string, len(items))

	for i, c := range items {
		ids[i] = fmt.Sprintf("%d", c.ID)

		_, _, display, unicode := localUtils.ResolveEmoji(c.SourceEmoji, bot.Client())

		emojiDisplay := c.SourceEmoji
		if !unicode {
			emojiDisplay = display
		}

		src := "Anywhere"
		if c.SourceChannelID != nil {
			src = fmt.Sprintf("<#%s>", *c.SourceChannelID)
		}

		emojiSources[i] = fmt.Sprintf("%s | %s", emojiDisplay, src)
		destinations[i] = fmt.Sprintf("<#%s>", c.TargetChannelID)
	}

	embed := discord.NewEmbed().
		WithColor(localStatic.EmbedColor).
		WithTitle(fmt.Sprintf("Starboards for %s", guildID)).
		WithFields(
			discord.EmbedField{Name: "ID", Value: strings.Join(ids, "\n"), Inline: boolPtr(true)},
			discord.EmbedField{Name: "Emoji | Source", Value: strings.Join(emojiSources, "\n"), Inline: boolPtr(true)},
			discord.EmbedField{Name: "Destination", Value: strings.Join(destinations, "\n"), Inline: boolPtr(true)},
		)

	if maxPage > 1 {
		embed = embed.WithEmbedFooter(&discord.EmbedFooter{
			Text: fmt.Sprintf("Page %d/%d", page, maxPage),
		})
	}

	var buttons []discord.InteractiveComponent
	if page > 1 {
		buttons = append(buttons, discord.NewPrimaryButton("◀️", fmt.Sprintf("/STARBOARD_LIST/%d", page-1)))
	}

	if page < maxPage {
		buttons = append(buttons, discord.NewPrimaryButton("▶️", fmt.Sprintf("/STARBOARD_LIST/%d", page+1)))
	}

	components := []discord.LayoutComponent{}
	if len(buttons) > 0 {
		components = append(components, discord.NewActionRow(buttons...))
	}

	return embed, components
}

func boolPtr(b bool) *bool { return &b }
