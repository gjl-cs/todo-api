package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal 统计 HTTP 请求次数。
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	// HTTPRequestDuration 统计 HTTP 请求耗时，单位为秒。
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	// CacheHitsTotal 统计缓存命中次数。
	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits.",
		},
	)

	// CacheMissesTotal 统计缓存未命中次数。
	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses.",
		},
	)

	// DBQueriesTotal 统计数据库查询次数。
	DBQueriesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries.",
		},
	)

	// RedirectsTotal 统计短链接跳转次数。
	RedirectsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "shortlink_redirects_total",
			Help: "Total number of short link redirects.",
		},
	)
)
