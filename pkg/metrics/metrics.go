package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registerOnce sync.Once

	workerUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gbs_worker_up",
		Help: "Whether the binlog worker is running.",
	})

	binlogEventsReadTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gbs_binlog_events_read_total",
			Help: "Total number of binlog events read from source.",
		},
		[]string{"event_type"},
	)

	binlogReadErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gbs_binlog_read_errors_total",
		Help: "Total number of source binlog read errors.",
	})

	replicationLagSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gbs_replication_lag_seconds",
		Help: "Current replication lag in seconds.",
	})

	eventsFilteredTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gbs_events_filtered_total",
			Help: "Total number of events skipped by filters.",
		},
		[]string{"kind"},
	)

	eventsAppliedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gbs_events_applied_total",
			Help: "Total number of events applied to destination.",
		},
		[]string{"event_type"},
	)

	eventApplyErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gbs_event_apply_errors_total",
			Help: "Total number of failures when processing events.",
		},
		[]string{"stage"},
	)

	eventApplyDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gbs_event_apply_duration_seconds",
			Help:    "Duration of applying an event to destination.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"event_type"},
	)

	stateCheckpointTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gbs_state_checkpoint_total",
			Help: "Total number of state checkpoint attempts.",
		},
		[]string{"result"},
	)

	stateLastCheckpointTimestampSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gbs_state_last_checkpoint_timestamp_seconds",
		Help: "Unix timestamp of the most recent successful checkpoint.",
	})
)

func ensureRegistered() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			workerUp,
			binlogEventsReadTotal,
			binlogReadErrorsTotal,
			replicationLagSeconds,
			eventsFilteredTotal,
			eventsAppliedTotal,
			eventApplyErrorsTotal,
			eventApplyDurationSeconds,
			stateCheckpointTotal,
			stateLastCheckpointTimestampSeconds,
		)
	})
}

func Handler() http.Handler {
	ensureRegistered()
	return promhttp.Handler()
}

func SetWorkerUp(up bool) {
	ensureRegistered()
	if up {
		workerUp.Set(1)
		return
	}
	workerUp.Set(0)
}

func IncBinlogEventsReadTotal(eventType string) {
	ensureRegistered()
	binlogEventsReadTotal.WithLabelValues(labelOrUnknown(eventType)).Inc()
}

func IncBinlogReadErrors() {
	ensureRegistered()
	binlogReadErrorsTotal.Inc()
}

func SetReplicationLag(seconds float64) {
	ensureRegistered()
	if seconds < 0 {
		seconds = 0
	}
	replicationLagSeconds.Set(seconds)
}

func IncEventsFiltered(kind string) {
	ensureRegistered()
	eventsFilteredTotal.WithLabelValues(labelOrUnknown(kind)).Inc()
}

func IncEventsApplied(eventType string) {
	ensureRegistered()
	eventsAppliedTotal.WithLabelValues(labelOrUnknown(eventType)).Inc()
}

func IncEventApplyErrors(stage string) {
	ensureRegistered()
	eventApplyErrorsTotal.WithLabelValues(labelOrUnknown(stage)).Inc()
}

func ObserveEventApplyDuration(eventType string, duration time.Duration) {
	ensureRegistered()
	eventApplyDurationSeconds.WithLabelValues(labelOrUnknown(eventType)).Observe(duration.Seconds())
}

func IncStateCheckpoint(result string) {
	ensureRegistered()
	stateCheckpointTotal.WithLabelValues(labelOrUnknown(result)).Inc()
}

func SetStateLastCheckpointTimestamp(at time.Time) {
	ensureRegistered()
	if at.IsZero() {
		return
	}
	stateLastCheckpointTimestampSeconds.Set(float64(at.Unix()))
}

func labelOrUnknown(v string) string {
	if v == "" {
		return "unknown"
	}
	return v
}
