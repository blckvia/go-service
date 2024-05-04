package prometheus

import "github.com/prometheus/client_golang/prometheus"

var CacheMissesTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "cacheMiss",
		Name:      "cache_request_total",
		Help:      "Total number of cache requests",
	},
	[]string{"goods_id", "project_id"},
)

var CacheHitsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "cacheHit",
		Name:      "cache_hits_total",
		Help:      "Total number of cache hits",
	},
	[]string{"goods_id", "project_id"},
)

var GoodsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "goodsCounter",
		Name:      "goods_counter",
		Help:      "Total requests for goods",
	},
	[]string{"project_id"},
)
