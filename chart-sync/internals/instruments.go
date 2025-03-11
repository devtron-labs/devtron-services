package internals

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var constLabels = map[string]string{"app": "chart-sync"}

var SyncOCIRepo = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name:        "sync_oci_repo",
		Help:        "no of update received in given chart and version",
		ConstLabels: constLabels,
	},
	[]string{})

var SyncRepo = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name:        "sync_repo",
		Help:        "no of update received in given chart and version",
		ConstLabels: constLabels,
	},
	[]string{})

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "http_duration_seconds",
		Help:        "Duration of HTTP requests.",
		ConstLabels: constLabels,
	}, []string{"path", "method", "status"})
)
