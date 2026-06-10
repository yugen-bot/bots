package guilds

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
)

func (m *GuildsModule) list(ctx *disgoplus.Ctx) {
	page := 1
	if v, ok := ctx.CommandData.OptInt("page"); ok {
		page = v
	}

	m.showList(ctx, page, false)
}

func (m *GuildsModule) listPage(ctx *disgoplus.Ctx) {
	page := 1

	if p, ok := ctx.MessageComponentOptions["page"]; ok {
		page64, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			page = 1
		} else {
			page = int(page64)
		}
	}

	m.showList(ctx, page, true)
}

func (m *GuildsModule) showList(
	ctx *disgoplus.Ctx,
	page int,
	isComponent bool,
) {
	if !isComponent {
		disgoplus.Defer(ctx, true)
	}

	guilds, total := m.guilds.GetData(page)

	if total == 0 {
		content := "There is no guild data available."
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
				discord.MessageCreate{
					Content: content,
					Flags:   discord.MessageFlagEphemeral,
				},
			)
		}

		return
	}

	if len(guilds) == 0 {
		content := fmt.Sprintf("No guilds found for page %d", page)
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
				discord.MessageCreate{
					Content: content,
					Flags:   discord.MessageFlagEphemeral,
				},
			)
		}

		return
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
		buttons = append(buttons, discord.NewPrimaryButton("◀️", fmt.Sprintf("ADMIN_GUILDS_LIST/%d", page-1)))
	}

	if page < maxPage {
		buttons = append(buttons, discord.NewPrimaryButton("▶️", fmt.Sprintf("ADMIN_GUILDS_LIST/%d", page+1)))
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
