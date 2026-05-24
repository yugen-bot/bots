package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var DictionaryCacheSize = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "dictionary_cache_size",
	Help: "Number of entries currently in the dictionary LRU cache",
})
