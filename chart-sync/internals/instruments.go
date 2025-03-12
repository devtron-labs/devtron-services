package internals

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var constLabels = map[string]string{"app": "chart-sync"}

// Counter metrics
var (
	SyncRepo = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "sync_repo",
			Help:        "Number of standard repository sync operations",
			ConstLabels: constLabels,
		},
		[]string{"repo_type", "repo_name"})

	RepoSyncErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "repo_sync_errors_total",
			Help:        "Total number of repository sync errors",
			ConstLabels: constLabels,
		},
		[]string{"repo_type", "error_type"})

	ChartVersionsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "chart_versions_processed_total",
			Help:        "Total number of chart versions processed successfully",
			ConstLabels: constLabels,
		},
		[]string{"repo_type", "chart_name"})

	ChartVersionsFailedProcessing = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "chart_versions_failed_processing_total",
			Help:        "Total number of chart versions that failed processing",
			ConstLabels: constLabels,
		},
		[]string{"repo_type", "chart_name", "error_type"})

	AppStoresCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name:        "app_stores_created_total",
			Help:        "Total number of app stores created during sync",
			ConstLabels: constLabels,
		})

	AppVersionsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name:        "app_versions_created_total",
			Help:        "Total number of app versions created during sync",
			ConstLabels: constLabels,
		})
)

// Histogram metrics
var (
	RepoSyncDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "repo_sync_duration_seconds",
			Help:        "Time taken to sync an entire repository",
			ConstLabels: constLabels,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"repo_type", "repo_name"})
)
