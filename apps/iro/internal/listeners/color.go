package listeners

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"github.com/zekroTJA/colorname"
	"github.com/zekroTJA/shinpuru/pkg/colors"
	"github.com/zekroTJA/timedmap"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

const (
	colorMatchesCap = 5
)

var rxColorHex = regexp.MustCompile(`^#?[\dA-Fa-f]{6,8}$`)

type ColorListener struct {
	bot        *discordgoplus.Bot
	cfg        *config.Config
	emojiCache *timedmap.TimedMap
}

func GetColorListener(container *di.Container) *ColorListener {
	utils.Logger.Info("Creating Color Listener")
	return &ColorListener{
		bot:        container.Get(static.DiBot).(*discordgoplus.Bot),
		cfg:        container.Get(static.DiConfig).(*config.Config),
		emojiCache: timedmap.New(1 * time.Minute),
	}
}

func AddColorListeners(container *di.Container) {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	colorListener := GetColorListener(container)
	bot.AddHandler(colorListener.MessageCreateHandler)
	bot.AddHandler(colorListener.MessageUpdateHandler)
	bot.AddHandler(colorListener.MessageReactionHandler)
}

func (listener *ColorListener) MessageCreateHandler(bot *discordgo.Session, event *discordgo.MessageCreate) {
	listener.process(bot, event.Message, false)
}

func (listener *ColorListener) MessageUpdateHandler(bot *discordgo.Session, event *discordgo.MessageUpdate) {
	listener.process(bot, event.Message, true)
}

func (listener *ColorListener) MessageReactionHandler(bot *discordgo.Session, event *discordgo.MessageReactionAdd) {
	self := listener.bot.State.User

	if event.UserID == self.ID {
		return
	}

	cacheKey := event.MessageID + event.Emoji.ID
	if !listener.emojiCache.Contains(cacheKey) {
		return
	}

	clr, ok := listener.emojiCache.GetValue(cacheKey).(*color.RGBA)
	if !ok {
		return
	}

	user, err := listener.bot.User(event.UserID)
	if err != nil {
		return
	}

	hexClr := colors.ToHex(clr)
	intClr := colors.ToInt(clr)
	cC, cM, cY, cK := color.RGBToCMYK(clr.R, clr.G, clr.B)
	yY, yCb, yCr := color.RGBToYCbCr(clr.R, clr.G, clr.B)

	colorName := "*could not be fetched*"
	matches := colorname.FindRGBA(clr)
	if len(matches) > 0 {
		precision := (1 - matches[0].AvgDiff/255) * 100
		colorName = fmt.Sprintf("**%s** *(%0.1f%%)*", matches[0].Name, precision)
	}

	emb := &discordgo.MessageEmbed{
		Color:       intClr,
		Title:       "#" + hexClr,
		Description: fmt.Sprintf("%s", colorName),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Hex",
				Value:  fmt.Sprintf("`#%s`", hexClr),
				Inline: true,
			},
			{
				Name:   "Int",
				Value:  fmt.Sprintf("`%d`", intClr),
				Inline: true,
			},
			{
				Name:  "RGBA",
				Value: fmt.Sprintf("`%03d, %03d, %03d, %03d`", clr.R, clr.G, clr.B, clr.A),
			},
			{
				Name:  "CMYK",
				Value: fmt.Sprintf("`%03d, %03d, %03d, %03d`", cC, cM, cY, cK),
			},
			{
				Name:  "YCbCr",
				Value: fmt.Sprintf("`%03d, %03d, %03d`", yY, yCb, yCr),
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Activated by " + user.String(),
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("https://singlecolorimage.com/get/%s/64x64", hexClr),
		},
	}

	_, err = bot.ChannelMessageSendComplex(event.ChannelID, &discordgo.MessageSend{
		Embed: emb,
		Reference: &discordgo.MessageReference{
			MessageID: event.MessageID,
			ChannelID: event.ChannelID,
			GuildID:   event.GuildID,
		},
	})
	if err != nil {
		utils.Logger.Info("Could not send embed message", err)
	}

	listener.emojiCache.Remove(cacheKey)
}

func (listener *ColorListener) process(bot *discordgo.Session, message *discordgo.Message, removeReactions bool) {
	if len(message.Content) < 6 {
		return
	}

	matches := make([]string, 0)

	content := strings.ReplaceAll(message.Content, "\n", " ")

	// Find color hex in message content using
	// predefined regex.
	for _, v := range strings.Split(content, " ") {
		if rxColorHex.MatchString(v) {
			matches = appendIfUnique(matches, v)
		}
	}

	cMatches := len(matches)

	// Cancel when no matches were found
	if cMatches == 0 {
		return
	}

	// Cap matches count to colorMatchesCap
	if cMatches > colorMatchesCap {
		matches = matches[:colorMatchesCap]
	}

	if removeReactions {
		if err := bot.MessageReactionsRemoveAll(message.ChannelID, message.ID); err != nil {
			utils.Logger.Info("Could not remove previous color reactions", err)
		}
	}

	// Execute reaction for each match
	for _, hexClr := range matches {
		listener.createReaction(bot, message, hexClr)
	}
}

func (listener *ColorListener) createReaction(bot *discordgo.Session, message *discordgo.Message, hexClr string) {
	// Remove trailing '#' from color code,
	// when existent
	hexClr = strings.TrimPrefix(hexClr, "#")

	// Parse hex color code to color.RGBA object
	clr, err := colors.FromHex(hexClr)
	if err != nil {
		utils.Logger.Info("Failed parsing color code", err)
		return
	}

	// Create a 24x24 px image with the parsed color
	// rendered as PNG into a buffer
	buff, err := colors.CreateImage(clr, 24, 24)
	if err != nil {
		utils.Logger.Info("Failed generating image data", err)
		return
	}

	// Encode the raw image data to a base64 string
	b64Data := base64.StdEncoding.EncodeToString(buff.Bytes())

	// Envelope the base64 data into data uri format
	dataUri := fmt.Sprintf("data:image/png;base64,%s", b64Data)

	// Upload guild emote
	clientId := listener.cfg.DiscordAppID
	emoji, err := bot.ApplicationEmojiCreate(clientId, &discordgo.EmojiParams{
		Name:  "hex" + hexClr,
		Image: dataUri,
	})
	if err != nil {
		utils.Logger.Info("Failed uploading emoji", err)
		return
	}

	// Delete the uploaded emote after 5 seconds
	// to give discords caching or whatever some
	// time to save the emoji.
	defer time.AfterFunc(5*time.Second, func() {
		if err = bot.ApplicationEmojiDelete(clientId, emoji.ID); err != nil {
			utils.Logger.Info("Failed deleting emoji", err)
		}
	})

	// Add reaction of the uploaded emote to the message
	err = bot.MessageReactionAdd(message.ChannelID, message.ID, emoji.APIName())
	if err != nil {
		utils.Logger.Info("Failed creating message reaction", err)
		return
	}

	// Set messageID + emojiID with RGBA color object
	// to emojiCache
	listener.emojiCache.Set(message.ID+emoji.ID, clr, 24*time.Hour)
}

// appendIfUnique appends the given elem to the
// passed slice only if the elem is not already
// contained in slice. Otherwise, slice will
// be returned unchanged.
func appendIfUnique(slice []string, elem string) []string {
	for _, m := range slice {
		if m == elem {
			return slice
		}
	}

	return append(slice, elem)
}
