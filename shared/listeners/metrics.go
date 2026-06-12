package listeners

import (
	"os"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/jurienhamaker/disgoplus"
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"github.com/shirou/gopsutil/v3/process"

	"jurien.dev/yugen/shared/metrics"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func reloadGauges(client *bot.Client) {
	time.Sleep(time.Second)
	metrics.TotalGuilds.Set(float64(client.Caches.GuildsLen()))
	metrics.TotalChannels.Set(float64(client.Caches.ChannelsLen()))
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

	metrics.ProcessCPUUsagePercentage.Set(getCPUPercentage(proc))

	if _, err := cron.AddFunc("@every 1m", func() {
		metrics.ProcessCPUUsagePercentage.Set(getCPUPercentage(proc))
	}); err != nil {
		panic(err)
	}
}

func AddMetricsListeners(container *di.Container) {
	disgoBot := container.Get(static.DiBot).(*disgoplus.Bot)
	client := disgoBot.Client()
	cron := container.Get(static.DiCron).(*cron.Cron)

	startCPUMetrics(cron)
	startLatencyCron(cron, client)
	registerMetricEventListeners(client)
}

func startLatencyCron(cron *cron.Cron, client *bot.Client) {
	writeLatency := func(g gateway.Gateway) {
		shardID := g.ShardID()
		shard := strconv.Itoa(shardID)
		if shardID < 0 {
			shard = "0"
		}
		metrics.DiscordLatency.WithLabelValues(shard).
			Set(float64(g.Latency().Milliseconds()))
	}

	if _, err := cron.AddFunc("@every 1m", func() {
		if client.HasShardManager() {
			for g := range client.ShardManager.Shards() {
				writeLatency(g)
			}
			return
		}
		if client.HasGateway() {
			writeLatency(client.Gateway)
		}
	}); err != nil {
		panic(err)
	}
}

func registerMetricEventListeners(client *bot.Client) {
	client.EventManager.AddEventListeners(
		bot.NewListenerFunc(func(e *events.Ready) {
			onReady(e, client)
		}),
		bot.NewListenerFunc(func(e *events.Resumed) {
			onResumed(e)
		}),
		bot.NewListenerFunc(func(e *events.GuildJoin) {
			go reloadGauges(client)
		}),
		bot.NewListenerFunc(func(e *events.GuildAvailable) {
			go reloadGauges(client)
		}),
		bot.NewListenerFunc(func(e *events.GuildLeave) {
			go reloadGauges(client)
		}),
		bot.NewListenerFunc(func(e *events.GuildUnavailable) {
			go reloadGauges(client)
		}),
		bot.NewListenerFunc(func(e *events.GuildChannelCreate) {
			go reloadGauges(client)
		}),
		bot.NewListenerFunc(func(e *events.GuildChannelDelete) {
			go reloadGauges(client)
		}),
		bot.NewListenerFunc(onSlashCommandInteraction),
		bot.NewListenerFunc(onComponentInteraction),
	)
}

func onReady(e *events.Ready, client *bot.Client) {
	metrics.DiscordConnected.Set(1)
	utils.Logger.Infof("Connected to Discord (shard %d)", e.ShardID())
	metrics.DiscordShards.Inc()

	go reloadGauges(client)

	metrics.TotalInteractions.Set(float64(utils.TotalRegisteredCommands()))
}

func onResumed(e *events.Resumed) {
	metrics.DiscordConnected.Set(1)
	utils.Logger.Infof("Resumed connection to Discord (shard %d)", e.ShardID())
}

func onSlashCommandInteraction(e *events.ApplicationCommandInteractionCreate) {
	if e.Data.Type() != discord.ApplicationCommandTypeSlash {
		return
	}

	data := e.Data.(discord.SlashCommandInteractionData)
	name := disgoplus.GetInteractionName(data)
	metrics.InteractionEventTotal.WithLabelValues("ChatInputCommandInteraction", name).
		Inc()
}

func onComponentInteraction(e *events.ComponentInteractionCreate) {
	customID := e.Data.CustomID()
	metrics.InteractionEventTotal.WithLabelValues("ButtonInteraction", customID).
		Inc()
}
