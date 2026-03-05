package state

import (
	"context"
	"errors"
	"time"

	"github.com/bobbyz3g/go-binlog-sync/pkg/metrics"
)

// Update describes a state update from a processed event.
type Update struct {
	Mode       Mode
	GTIDSet    string
	BinlogFile string
	BinlogPos  uint64
}

// Recorder batches and persists state updates.
type Recorder struct {
	store Store
	st    State
	every int
	since int
	dirty bool
}

func NewRecorder(store Store, st State, every int) *Recorder {
	if every <= 0 {
		every = 1
	}
	return &Recorder{
		store: store,
		st:    st,
		every: every,
	}
}

func (r *Recorder) Record(ctx context.Context, upd Update) error {
	if r == nil {
		return nil
	}
	if r.store == nil {
		return errors.New("state store is nil")
	}
	if upd.Mode != "" {
		r.st.Mode = upd.Mode
	}
	if upd.GTIDSet != "" {
		r.st.GTIDSet = upd.GTIDSet
	}
	if upd.BinlogFile != "" {
		r.st.BinlogFile = upd.BinlogFile
	}
	if upd.BinlogPos > 0 {
		r.st.BinlogPos = upd.BinlogPos
	}
	if upd.GTIDSet != "" || upd.BinlogFile != "" || upd.BinlogPos > 0 || upd.Mode != "" {
		r.dirty = true
	}
	if r.every <= 0 {
		r.every = 1
	}
	r.since++
	if r.since >= r.every {
		r.since = 0
		return r.save(ctx)
	}
	return nil
}

func (r *Recorder) Flush(ctx context.Context) error {
	if r == nil || !r.dirty {
		return nil
	}
	return r.save(ctx)
}

func (r *Recorder) save(ctx context.Context) error {
	if r.store == nil {
		return errors.New("state store is nil")
	}
	if err := r.store.Save(ctx, &r.st); err != nil {
		metrics.IncStateCheckpoint("fail")
		return err
	}
	metrics.IncStateCheckpoint("success")
	metrics.SetStateLastCheckpointTimestamp(time.Now())
	r.dirty = false
	return nil
}

func (r *Recorder) State() State {
	return r.st
}
