package list

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	localStatic "jurien.dev/yugen/hoshi/internal/static"
	localUtils "jurien.dev/yugen/hoshi/internal/utils"
	"jurien.dev/yugen/shared/static"
)

func (m *ListModule) list(ctx *disgoplus.Ctx) {
	page := 1
	if v, ok := ctx.CommandData.OptInt("page"); ok {
		page = v
	}

	m.showList(ctx, page, false)
}

func (m *ListModule) listPage(ctx *disgoplus.Ctx) {
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

func (m *ListModule) showList(
	ctx *disgoplus.Ctx,
	page int,
	isComponent bool,
) {
	if !isComponent {
		disgoplus.Defer(ctx, true)
	}

	bot := m.container.Get(static.DiClient).(*disgoplus.Bot)

	items, total, err := m.starboard.GetStarboards(
		context.Background(),
		ctx.GuildID.String(),
		page,
	)
	if err != nil {
		if isComponent {
			disgoplus.MessageComponentError(ctx)
		} else {
			disgoplus.InteractionError(ctx, true)
		}

		return
	}

	if total == 0 {
		content := "No starboards have been configured yet."
		if isComponent {
			empty := []discord.Embed{}
			emptyComponents := []discord.LayoutComponent{}
			disgoplus.Update(
				ctx,
				discord.MessageUpdate{
					Content:    &content,
					Embeds:     &empty,
					Components: &emptyComponents,
				},
			)
		} else {
			disgoplus.FollowUp(
				ctx,
				discord.MessageCreate{Content: content, Flags: discord.MessageFlagEphemeral},
			)
		}

		return
	}

	if len(items) == 0 {
		content := fmt.Sprintf("No starboards found for page %d", page)
		if isComponent {
			empty := []discord.Embed{}
			emptyComponents := []discord.LayoutComponent{}
			disgoplus.Update(
				ctx,
				discord.MessageUpdate{
					Content:    &content,
					Embeds:     &empty,
					Components: &emptyComponents,
				},
			)
		} else {
			disgoplus.FollowUp(
				ctx,
				discord.MessageCreate{Content: content, Flags: discord.MessageFlagEphemeral},
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
		WithTitle(fmt.Sprintf("Starboards for %s", ctx.GuildID.String())).
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
		buttons = append(buttons, discord.NewPrimaryButton("◀️", fmt.Sprintf("STARBOARD_LIST/%d", page-1)))
	}

	if page < maxPage {
		buttons = append(buttons, discord.NewPrimaryButton("▶️", fmt.Sprintf("STARBOARD_LIST/%d", page+1)))
	}

	components := []discord.LayoutComponent{}
	if len(buttons) > 0 {
		components = append(components, discord.NewActionRow(buttons...))
	}

	if isComponent {
		embeds := []discord.Embed{embed}
		disgoplus.Update(ctx, discord.MessageUpdate{
			Embeds:     &embeds,
			Components: &components,
		})
	} else {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Embeds:     []discord.Embed{embed},
			Components: components,
			Flags:      discord.MessageFlagEphemeral,
		})
	}
}

func boolPtr(b bool) *bool { return &b }
