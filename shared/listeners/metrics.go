package listeners

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/jurienhamaker/disgoplus"
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"github.com/shirou/gopsutil/v3/process"

	"jurien.dev/yugen/shared/metrics"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

// lastHeartbeatTime stores the last heartbeat ACK timestamp per shard for
// approximate reconnect-duration logging (no explicit disconnect event in disgo).
var lastHeartbeatTime sync.Map // shardID (int) → time.Time

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
	startLatencyCron(cron)
	registerMetricEventListeners(client)
}

func startLatencyCron(cron *cron.Cron) {
	if _, err := cron.AddFunc("@every 1m", func() {
		lastHeartbeatTime.Range(func(k, v any) bool {
			shardID := k.(int)
			t := v.(time.Time)

			shard := strconv.Itoa(shardID)
			if shardID < 0 {
				shard = "0"
			}

			elapsed := time.Since(t)
			metrics.DiscordLatency.WithLabelValues(shard).
				Set(float64(elapsed.Milliseconds()))

			return true
		})
	}); err != nil {
		panic(err)
	}
}

func registerMetricEventListeners(client *bot.Client) {
	client.EventManager.AddEventListeners(
		bot.NewListenerFunc(onHeartbeatAck),
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

func onHeartbeatAck(e *events.HeartbeatAck) {
	lastHeartbeatTime.Store(e.ShardID(), time.Now())
	latency := e.NewHeartbeat.Sub(e.LastHeartbeat)

	shard := strconv.Itoa(e.ShardID())
	if e.ShardID() < 0 {
		shard = "0"
	}

	metrics.DiscordLatency.WithLabelValues(shard).
		Set(float64(latency.Milliseconds()))
}

func onReady(e *events.Ready, client *bot.Client) {
	shardID := e.ShardID()

	metrics.DiscordConnected.Set(1)

	if lh, ok := lastHeartbeatTime.Load(shardID); ok {
		gap := time.Since(lh.(time.Time))
		if gap > 10*time.Second {
			utils.Logger.Infof(
				"Reconnected to Discord (shard %d) after ~%s",
				shardID,
				gap.Round(time.Second),
			)
		} else {
			utils.Logger.Infof("Connected to Discord (shard %d)", shardID)
		}
	} else {
		utils.Logger.Infof("Connected to Discord (shard %d)", shardID)
	}

	metrics.DiscordShards.Inc()

	go reloadGauges(client)

	metrics.TotalInteractions.Set(float64(utils.TotalRegisteredCommands()))
}

func onResumed(e *events.Resumed) {
	shardID := e.ShardID()

	metrics.DiscordConnected.Set(1)

	if lh, ok := lastHeartbeatTime.Load(shardID); ok {
		gap := time.Since(lh.(time.Time))
		utils.Logger.Infof(
			"Reconnected to Discord (shard %d) after ~%s",
			shardID,
			gap.Round(time.Second),
		)
	} else {
		utils.Logger.Infof("Resumed connection to Discord (shard %d)", shardID)
	}
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
