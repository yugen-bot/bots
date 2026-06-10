package listeners

import (
	"fmt"
	"image/color"
	"regexp"
	"strings"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"
	"github.com/zekroTJA/colorname"
	"github.com/zekroTJA/shinpuru/pkg/colors"
	"github.com/zekroTJA/timedmap"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

const colorMatchesCap = 5

var rxColorHex = regexp.MustCompile(`^#?[\dA-Fa-f]{6,8}$`)

type ColorListener struct {
	client     *bot.Client
	cfg        *config.Config
	emojiCache *timedmap.TimedMap
}

func GetColorListener(container *di.Container) *ColorListener {
	utils.Logger.Info("Creating Color Listener")

	return &ColorListener{
		client:     container.Get(static.DiBot).(*disgoplus.Bot).Client(),
		cfg:        container.Get(static.DiConfig).(*config.Config),
		emojiCache: timedmap.New(1 * time.Minute),
	}
}

func AddColorListeners(container *di.Container) {
	disgoBot := container.Get(static.DiBot).(*disgoplus.Bot)
	cl := GetColorListener(container)

	disgoBot.Client().EventManager.AddEventListeners(
		bot.NewListenerFunc(func(e *events.MessageCreate) {
			if e.GuildID == nil {
				return
			}

			cl.process(e.Message, false)
		}),

		bot.NewListenerFunc(func(e *events.MessageUpdate) {
			if e.GuildID == nil {
				return
			}

			cl.process(e.Message, true)
		}),

		bot.NewListenerFunc(func(e *events.GuildMessageReactionAdd) {
			self, ok := cl.client.Caches.SelfUser()
			if !ok || e.UserID == self.ID {
				return
			}

			var emojiID string
			if e.Emoji.ID != nil {
				emojiID = e.Emoji.ID.String()
			}

			cacheKey := e.MessageID.String() + emojiID
			if !cl.emojiCache.Contains(cacheKey) {
				return
			}

			clr, ok := cl.emojiCache.GetValue(cacheKey).(*color.RGBA)
			if !ok {
				return
			}

			user, err := cl.client.Rest.GetUser(e.UserID)
			if err != nil {
				return
			}

			hexClr := colors.ToHex(clr)
			intClr := colors.ToInt(clr)
			cC, cM, cY, cK := color.RGBToCMYK(clr.R, clr.G, clr.B)
			yY, yCb, yCr := color.RGBToYCbCr(clr.R, clr.G, clr.B)

			colorNameStr := "*could not be fetched*"

			matches := colorname.FindRGBA(clr)
			if len(matches) > 0 {
				precision := (1 - matches[0].AvgDiff/255) * 100
				colorNameStr = fmt.Sprintf(
					"**%s** *(%0.1f%%)*",
					matches[0].Name,
					precision,
				)
			}

			emb := discord.NewEmbed().
				WithColor(intClr).
				WithTitle("#"+hexClr).
				WithDescription(colorNameStr).
				WithFields(
					discord.EmbedField{
						Name:   "Hex",
						Value:  fmt.Sprintf("`#%s`", hexClr),
						Inline: boolPtr(true),
					},
					discord.EmbedField{
						Name:   "Int",
						Value:  fmt.Sprintf("`%d`", intClr),
						Inline: boolPtr(true),
					},
					discord.EmbedField{
						Name:  "RGBA",
						Value: fmt.Sprintf("`%03d, %03d, %03d, %03d`", clr.R, clr.G, clr.B, clr.A),
					},
					discord.EmbedField{
						Name:  "CMYK",
						Value: fmt.Sprintf("`%03d, %03d, %03d, %03d`", cC, cM, cY, cK),
					},
					discord.EmbedField{
						Name:  "YCbCr",
						Value: fmt.Sprintf("`%03d, %03d, %03d`", yY, yCb, yCr),
					},
				).
				WithFooterText("Activated by " + user.Username).
				WithThumbnail(fmt.Sprintf("https://singlecolorimage.com/get/%s/64x64", hexClr))

			msg := discord.NewMessageCreate().
				AddEmbeds(emb).
				WithMessageReferenceByID(e.MessageID)

			if _, err := cl.client.Rest.CreateMessage(
				e.ChannelID,
				msg,
			); err != nil {
				utils.Logger.Info("Could not send embed message", err)
			}

			cl.emojiCache.Remove(cacheKey)
		}),
	)
}

func (l *ColorListener) process(message discord.Message, removeReactions bool) {
	if len(message.Content) < 6 {
		return
	}

	matches := make([]string, 0)

	content := strings.ReplaceAll(message.Content, "\n", " ")
	for _, v := range strings.Split(content, " ") {
		if rxColorHex.MatchString(v) {
			matches = appendIfUnique(matches, v)
		}
	}

	if len(matches) == 0 {
		return
	}

	if len(matches) > colorMatchesCap {
		matches = matches[:colorMatchesCap]
	}

	if removeReactions {
		if err := l.client.Rest.RemoveAllReactions(
			message.ChannelID,
			message.ID,
		); err != nil {
			utils.Logger.Info("Could not remove previous color reactions", err)
		}
	}

	for _, hexClr := range matches {
		l.createReaction(message, hexClr)
	}
}

func (l *ColorListener) createReaction(message discord.Message, hexClr string) {
	hexClr = strings.TrimPrefix(hexClr, "#")

	clr, err := colors.FromHex(hexClr)
	if err != nil {
		utils.Logger.Info("Failed parsing color code", err)
		return
	}

	buff, err := colors.CreateImage(clr, 24, 24)
	if err != nil {
		utils.Logger.Info("Failed generating image data", err)
		return
	}

	icon := discord.NewIconRaw(discord.IconTypePNG, buff.Bytes())

	appID, err := snowflake.Parse(l.cfg.DiscordAppID)
	if err != nil {
		utils.Logger.Info("Failed parsing app ID", err)
		return
	}

	emoji, err := l.client.Rest.CreateApplicationEmoji(
		appID,
		discord.EmojiCreate{
			Name:  "hex" + hexClr,
			Image: *icon,
		},
	)
	if err != nil {
		utils.Logger.Info("Failed uploading emoji", err)
		return
	}

	defer time.AfterFunc(5*time.Second, func() {
		if err := l.client.Rest.DeleteApplicationEmoji(
			appID,
			emoji.ID,
		); err != nil {
			utils.Logger.Info("Failed deleting emoji", err)
		}
	})

	if err := l.client.Rest.AddReaction(
		message.ChannelID,
		message.ID,
		emoji.Reaction(),
	); err != nil {
		utils.Logger.Info("Failed creating message reaction", err)
		return
	}

	l.emojiCache.Set(message.ID.String()+emoji.ID.String(), clr, 24*time.Hour)
}

func boolPtr(b bool) *bool { return &b }

func appendIfUnique(slice []string, elem string) []string {
	for _, m := range slice {
		if m == elem {
			return slice
		}
	}

	return append(slice, elem)
}
