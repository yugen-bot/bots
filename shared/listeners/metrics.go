package listeners

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/metrics"
	"jurien.dev/yugen/shared/static"
)

func setLatency(bot *discordgoplus.Bot) {
	latency := bot.HeartbeatLatency()
	metrics.DiscordLatency.Set(float64(latency.Milliseconds()))
}

func reloadGuilds(bot *discordgoplus.Bot) {
	time.Sleep(time.Second)

	guilds := bot.State.Guilds
	metrics.TotalGuilds.Set(float64(len(guilds)))
}

func reloadChannels(bot *discordgoplus.Bot) {
	time.Sleep(time.Second)

	guilds := bot.State.Guilds

	channelsLen := 0
	for _, guild := range guilds {
		channelsLen = channelsLen + len(guild.Channels)
	}

	metrics.TotalChannels.Set(float64(channelsLen))
}

func reloadInteractions(bot *discordgoplus.Bot) {
	time.Sleep(time.Second)

	interactionsLen := 0

	for _, command := range bot.Router.Commands {
		if command.SubCommands != nil {
			interactionsLen = interactionsLen + len(command.SubCommands.Commands)
			continue
		}

		interactionsLen = interactionsLen + 1
	}

	metrics.TotalInteractions.Set(float64(interactionsLen))
}

func reloadGuages(bot *discordgoplus.Bot) {
	go reloadGuilds(bot)
	go reloadChannels(bot)
	go reloadInteractions(bot)
}

func AddMetricsListeners(container *di.Container) {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)
	cron := container.Get(static.DiCron).(*cron.Cron)

	setLatency(bot)
	cron.AddFunc("@every 1m", func() {
		go setLatency(bot)
	})

	bot.AddHandler(func(session *discordgo.Session, event *discordgo.Ready) {
		metrics.DiscordConnected.Set(1)

		go reloadGuages(bot)
	})

	bot.AddHandler(func(session *discordgo.Session, event *discordgo.Disconnect) {
		metrics.DiscordConnected.Set(0)
	})

	bot.AddHandler(func(session *discordgo.Session, event *discordgo.GuildCreate) {
		go reloadGuilds(bot)
	})

	bot.AddHandler(func(session *discordgo.Session, event *discordgo.GuildDelete) {
		go reloadGuilds(bot)
	})

	bot.AddHandler(func(session *discordgo.Session, event *discordgo.ChannelCreate) {
		go reloadChannels(bot)
	})

	bot.AddHandler(func(session *discordgo.Session, event *discordgo.ChannelDelete) {
		go reloadChannels(bot)
	})

	bot.AddHandler(func(session *discordgo.Session, event *discordgo.InteractionCreate) {
		if event.Type != discordgo.InteractionApplicationCommand {
			return
		}

		data := event.ApplicationCommandData()
		name := discordgoplus.GetInteractionName(&data)

		metrics.InteractionEventTotal.WithLabelValues("ChatInputCommandInteraction", name).Inc()
	})

	bot.AddHandler(func(bot *discordgo.Session, event *discordgo.InteractionCreate) {
		if event.Type != discordgo.InteractionMessageComponent {
			return
		}

		data := event.MessageComponentData()
		metrics.InteractionEventTotal.WithLabelValues("ButtonInteraction", data.CustomID).Inc()
	})
}
