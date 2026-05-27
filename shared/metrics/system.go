package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var ProcessCPUUsagePercentage = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "cpu_usage_percent",
	Help: "The amount of CPU used by the bot",
})
