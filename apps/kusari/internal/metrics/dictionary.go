package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var DictionaryCacheHits = promauto.NewCounter(prometheus.CounterOpts{
	Name: "dictionary_cache_hits_total",
	Help: "Total number of dictionary cache hits",
})

var DictionaryCacheMisses = promauto.NewCounter(prometheus.CounterOpts{
	Name: "dictionary_cache_misses_total",
	Help: "Total number of dictionary cache misses",
})
