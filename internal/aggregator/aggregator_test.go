package aggregator

import (
	"sync"
	"testing"
	"time"
)

type recorder struct {
	mu   sync.Mutex
	got  []*Bucket
}

func (r *recorder) flush(b *Bucket) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.got = append(r.got, b)
}

func (r *recorder) snapshot() []*Bucket {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*Bucket, len(r.got))
	copy(out, r.got)
	return out
}

func TestImmediateMode_EmitsEachObservation(t *testing.T) {
	var r recorder
	a := New(0, r.flush)
	defer a.Close()

	a.Observe("k1", -50, func() interface{} { return "rec1" })
	a.Observe("k1", -52, func() interface{} { return "rec2" })
	a.Observe("k2", -60, func() interface{} { return "rec3" })

	got := r.snapshot()
	if len(got) != 3 {
		t.Fatalf("immediate mode: got %d flushes, want 3", len(got))
	}
	for i, b := range got {
		if len(b.RSSI) != 1 {
			t.Errorf("flush %d: want single sample, got %d", i, len(b.RSSI))
		}
	}
}

func TestWindowMode_CollapsesSameKey(t *testing.T) {
	var r recorder
	a := New(80*time.Millisecond, r.flush)
	defer a.Close()

	// Three samples for k1, one for k2 — all within window.
	a.Observe("k1", -50, func() interface{} { return "rec1" })
	a.Observe("k1", -55, func() interface{} { return "(should-not-be-built)" })
	a.Observe("k1", -52, func() interface{} { return "(should-not-be-built)" })
	a.Observe("k2", -60, func() interface{} { return "rec2" })

	// Wait for the window to expire and the run-loop ticker to fire.
	time.Sleep(2 * time.Second)

	got := r.snapshot()
	if len(got) != 2 {
		t.Fatalf("want 2 buckets (one per key), got %d", len(got))
	}

	byKey := map[string]*Bucket{}
	for _, b := range got {
		byKey[b.Key] = b
	}
	if k1 := byKey["k1"]; k1 == nil || len(k1.RSSI) != 3 {
		t.Errorf("k1: want 3 rssi samples, got %v", k1)
	}
	if k2 := byKey["k2"]; k2 == nil || len(k2.RSSI) != 1 {
		t.Errorf("k2: want 1 rssi sample, got %v", k2)
	}
}

func TestWindowMode_RecordOnlyBuiltOnce(t *testing.T) {
	var r recorder
	a := New(60*time.Millisecond, r.flush)
	defer a.Close()

	builtTimes := 0
	mkRecord := func() interface{} {
		builtTimes++
		return "the only record"
	}
	a.Observe("k", 1, mkRecord)
	a.Observe("k", 2, mkRecord)
	a.Observe("k", 3, mkRecord)

	if builtTimes != 1 {
		t.Errorf("record built %d times; want exactly 1", builtTimes)
	}
}

func TestObserveExisting_SkipsUnknownKey(t *testing.T) {
	var r recorder
	a := New(time.Hour, r.flush)

	// No bucket yet → false, and nothing created (the caller must parse + Observe).
	if a.ObserveExisting("k", -50) {
		t.Fatal("ObserveExisting on unknown key returned true")
	}

	// First real sighting builds the record exactly once.
	builtTimes := 0
	a.Observe("k", -50, func() interface{} { builtTimes++; return "rec" })

	// Now the bucket exists → samples merge without rebuilding the record.
	if !a.ObserveExisting("k", -55) {
		t.Fatal("ObserveExisting on known key returned false")
	}
	if !a.ObserveExisting("k", -52) {
		t.Fatal("ObserveExisting on known key returned false")
	}
	if builtTimes != 1 {
		t.Errorf("record built %d times; want exactly 1", builtTimes)
	}

	a.Close()
	got := r.snapshot()
	if len(got) != 1 {
		t.Fatalf("want 1 bucket, got %d", len(got))
	}
	if len(got[0].RSSI) != 3 {
		t.Errorf("want 3 rssi samples (1 Observe + 2 ObserveExisting), got %d", len(got[0].RSSI))
	}
}

func TestObserveExisting_ImmediateModeAlwaysFalse(t *testing.T) {
	var r recorder
	a := New(0, r.flush)
	defer a.Close()

	// Immediate mode keeps no persistent buckets, so the short-circuit must
	// never engage — every packet has to be parsed and emitted.
	a.Observe("k", -50, func() interface{} { return "rec" })
	if a.ObserveExisting("k", -55) {
		t.Fatal("ObserveExisting must always be false in immediate mode")
	}
}

func TestClose_FlushesPending(t *testing.T) {
	var r recorder
	a := New(time.Hour, r.flush)
	a.Observe("k1", -50, func() interface{} { return "rec" })
	a.Observe("k2", -60, func() interface{} { return "rec" })
	a.Close()

	got := r.snapshot()
	if len(got) != 2 {
		t.Errorf("Close did not flush; got %d", len(got))
	}
}
