package listeners

import (
	"context"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type ReactionListener struct {
	client *bot.Client
	svc    *services.StarboardService
	timers sync.Map
}

func GetReactionListener(container *di.Container) *ReactionListener {
	return &ReactionListener{
		client: container.Get(sharedStatic.DiClient).(*disgoplus.Bot).Client(),
		svc:    container.Get(localStatic.DiStarboard).(*services.StarboardService),
	}
}

func AddReactionListeners(container *di.Container) {
	l := GetReactionListener(container)
	disgoBot := container.Get(sharedStatic.DiClient).(*disgoplus.Bot)
	disgoBot.Client().EventManager.AddEventListeners(
		bot.NewListenerFunc(l.OnMessageReactionAdd),
		bot.NewListenerFunc(l.OnMessageReactionRemove),
	)
}

func (l *ReactionListener) OnMessageReactionAdd(e *events.GuildMessageReactionAdd) {
	self, ok := l.client.Caches.SelfUser()
	if ok && e.UserID == self.ID {
		return
	}

	var emojiID string
	if e.Emoji.ID != nil {
		emojiID = e.Emoji.ID.String()
	}

	emojiName := ""
	if e.Emoji.Name != nil {
		emojiName = *e.Emoji.Name
	}

	l.debounce(e.ChannelID.String(), e.MessageID.String(), e.GuildID.String(), emojiID, emojiName)
}

func (l *ReactionListener) OnMessageReactionRemove(e *events.GuildMessageReactionRemove) {
	var emojiID string
	if e.Emoji.ID != nil {
		emojiID = e.Emoji.ID.String()
	}

	emojiName := ""
	if e.Emoji.Name != nil {
		emojiName = *e.Emoji.Name
	}

	l.debounce(e.ChannelID.String(), e.MessageID.String(), e.GuildID.String(), emojiID, emojiName)
}

func (l *ReactionListener) debounce(channelID, messageID, guildID, emojiID, emojiName string) {
	if existing, ok := l.timers.LoadAndDelete(messageID); ok {
		existing.(*time.Timer).Stop()
	}

	t := time.AfterFunc(250*time.Millisecond, func() {
		l.timers.Delete(messageID)
		l.svc.CheckReaction(context.Background(), channelID, messageID, guildID, emojiID, emojiName)
	})

	l.timers.Store(messageID, t)
}
