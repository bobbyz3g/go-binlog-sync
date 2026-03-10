package state

import (
	"testing"
	"time"
)

func TestMySQLTimestampValue(t *testing.T) {
	ts := time.Date(2026, time.March, 10, 20, 55, 19, 987654321, time.FixedZone("UTC+8", 8*60*60))

	got := mysqlTimestampValue(ts)

	if got != "2026-03-10 12:55:19" {
		t.Fatalf("expected UTC timestamp string, got %q", got)
	}
}

func TestMySQLTimestampValueZero(t *testing.T) {
	if got := mysqlTimestampValue(time.Time{}); got != "" {
		t.Fatalf("expected empty string for zero time, got %q", got)
	}
}
