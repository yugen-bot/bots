package listeners

import (
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"github.com/shirou/gopsutil/v3/process"
	"jurien.dev/yugen/shared/metrics"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func setLatency(bot *discordgoplus.Bot) {
	bot.Each(func(b *discordgoplus.Bot) {
		shard := strconv.Itoa(b.ShardID)
		if b.ShardID < 0 {
			shard = "0"
		}

		metrics.DiscordLatency.WithLabelValues(shard).
			Set(float64(b.HeartbeatLatency().Milliseconds()))
	})
}

func reloadGuilds(bot *discordgoplus.Bot) {
	time.Sleep(time.Second)

	var total int64

	bot.Each(func(b *discordgoplus.Bot) {
		atomic.AddInt64(&total, int64(len(b.State.Guilds)))
	})
	metrics.TotalGuilds.Set(float64(atomic.LoadInt64(&total)))
}

func reloadChannels(bot *discordgoplus.Bot) {
	time.Sleep(time.Second)

	var total int64

	bot.Each(func(b *discordgoplus.Bot) {
		var n int64
		for _, g := range b.State.Guilds {
			n += int64(len(g.Channels))
		}

		atomic.AddInt64(&total, n)
	})
	metrics.TotalChannels.Set(float64(atomic.LoadInt64(&total)))
}

func reloadInteractions(bot *discordgoplus.Bot) {
	time.Sleep(time.Second)

	interactionsLen := 0

	for _, command := range bot.Router.Commands {
		if command.SubCommands != nil {
			interactionsLen += len(command.SubCommands.Commands)
			continue
		}

		interactionsLen++
	}

	metrics.TotalInteractions.Set(float64(interactionsLen))
}

func reloadGuages(bot *discordgoplus.Bot) {
	go reloadGuilds(bot)
	go reloadChannels(bot)
	go reloadInteractions(bot)
}

func getCPUPercentage(proc *process.Process) float64 {
	if pct, err := proc.Percent(0); err == nil {
		return pct
	}

	return 0
}

func startCPUMetrics(cron *cron.Cron) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		panic("metrics: cannot create process handle: " + err.Error())
	}

	pct := getCPUPercentage(proc)
	metrics.ProcessCPUUsagePercentage.Set(pct)

	if _, err := cron.AddFunc("@every 1m", func() {
		pct := getCPUPercentage(proc)
		metrics.ProcessCPUUsagePercentage.Set(pct)
	}); err != nil {
		panic(err)
	}
}

func AddMetricsListeners(container *di.Container) {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)
	cron := container.Get(static.DiCron).(*cron.Cron)

	startCPUMetrics(cron)

	setLatency(bot)
	if _, err := cron.AddFunc("@every 1m", func() {
		go setLatency(bot)
	}); err != nil {
		panic(err)
	}

	shards := 0
	bot.AddHandler(func(session *discordgo.Session, event *discordgo.Ready) {
		if bot.Sharded {
			shards = shards + 1
		}

		if !bot.Sharded {
			shards = 1
		}

		metrics.DiscordShards.Set(float64(shards))

		go reloadGuages(bot)
	})

	bot.AddHandler(
		func(session *discordgo.Session, event *discordgo.Connect) {
			utils.Logger.Info("Connected to Discord")
			metrics.DiscordConnected.Set(1)
		},
	)

	bot.AddHandler(
		func(session *discordgo.Session, event *discordgo.Disconnect) {
			utils.Logger.Info("Disconnected from Discord")
			metrics.DiscordConnected.Set(0)
		},
	)

	bot.AddHandler(
		func(session *discordgo.Session, event *discordgo.GuildCreate) {
			go reloadGuilds(bot)
		},
	)

	bot.AddHandler(
		func(session *discordgo.Session, event *discordgo.GuildDelete) {
			go reloadGuilds(bot)
		},
	)

	bot.AddHandler(
		func(session *discordgo.Session, event *discordgo.ChannelCreate) {
			go reloadChannels(bot)
		},
	)

	bot.AddHandler(
		func(session *discordgo.Session, event *discordgo.ChannelDelete) {
			go reloadChannels(bot)
		},
	)

	bot.AddHandler(
		func(session *discordgo.Session, event *discordgo.InteractionCreate) {
			if event.Type != discordgo.InteractionApplicationCommand {
				return
			}

			data := event.ApplicationCommandData()
			name := discordgoplus.GetInteractionName(&data)

			metrics.InteractionEventTotal.WithLabelValues("ChatInputCommandInteraction", name).
				Inc()
		},
	)

	bot.AddHandler(
		func(bot *discordgo.Session, event *discordgo.InteractionCreate) {
			if event.Type != discordgo.InteractionMessageComponent {
				return
			}

			data := event.MessageComponentData()
			metrics.InteractionEventTotal.WithLabelValues("ButtonInteraction", data.CustomID).
				Inc()
		},
	)
}
