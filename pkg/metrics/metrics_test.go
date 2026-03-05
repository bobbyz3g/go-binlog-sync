package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricsRecordedAndGathered(t *testing.T) {
	SetWorkerUp(true)
	IncBinlogEventsReadTotal("query_event")
	IncBinlogReadErrors()
	SetReplicationLag(2.5)
	IncEventsFiltered("rows")
	IncEventsApplied("query")
	IncEventApplyErrors("exec")
	ObserveEventApplyDuration("query", 25*time.Millisecond)
	IncStateCheckpoint("success")
	SetStateLastCheckpointTimestamp(time.Unix(1700000000, 0))

	got, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	want := map[string]bool{
		"gbs_worker_up":                               false,
		"gbs_binlog_events_read_total":                false,
		"gbs_binlog_read_errors_total":                false,
		"gbs_replication_lag_seconds":                 false,
		"gbs_events_filtered_total":                   false,
		"gbs_events_applied_total":                    false,
		"gbs_event_apply_errors_total":                false,
		"gbs_event_apply_duration_seconds":            false,
		"gbs_state_checkpoint_total":                  false,
		"gbs_state_last_checkpoint_timestamp_seconds": false,
	}

	for _, mf := range got {
		name := mf.GetName()
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}

	for name, found := range want {
		if !found {
			t.Fatalf("expected metric %s to be gathered", name)
		}
	}
}
