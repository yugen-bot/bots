package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var DiscordConnected = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "discord_connected",
	Help: "Determines if the bot is connected to Discord",
})

var DiscordLatency = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "discord_latency",
	Help: "Latency to Discord (ms) per shard",
}, []string{"shard"})
