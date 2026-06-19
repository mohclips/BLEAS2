// Package aggregator collects observations of the same logical BLE
// advertisement across a rolling time window and emits one summary per
// (key, window).
//
// Unlike a pure dedup, the aggregator preserves every RSSI sample seen
// during the window — useful for motion / proximity inference — and reports
// the count of observations. The first observation's parsed record is kept
// as the template; subsequent observations only update the timing and RSSI
// vectors. When the window is zero, the aggregator degenerates to immediate
// per-observation emission (count=1, single rssi sample).
package aggregator

import (
	"sort"
	"sync"
	"time"
)

// Bucket is the snapshot the flush callback receives. The Record holds the
// caller-supplied parsed payload (typed by the caller; the aggregator does
// not interpret it).
type Bucket struct {
	Key       string
	FirstSeen time.Time
	LastSeen  time.Time
	RSSI      []int
	Record    interface{}
}

// FlushFunc is invoked exactly once per bucket: either when the window
// expires (batched mode), at Close, or synchronously per call (immediate
// mode when window <= 0).
type FlushFunc func(*Bucket)

// Aggregator owns its own goroutine that scans for expired buckets at a
// fraction of the window. Close stops the loop and flushes any in-flight
// buckets.
type Aggregator struct {
	window time.Duration
	flush  FlushFunc

	mu      sync.Mutex
	buckets map[string]*Bucket

	stop chan struct{}
	done chan struct{}
}

// New creates an aggregator with the given window. window <= 0 disables
// batching: each Observe immediately invokes flush.
func New(window time.Duration, flush FlushFunc) *Aggregator {
	a := &Aggregator{
		window:  window,
		flush:   flush,
		buckets: make(map[string]*Bucket),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
	go a.run()
	return a
}

// Observe records a single RSSI sample for key. recordFn is only invoked on
// the FIRST observation in the current window — that's intentional: the
// emitted record reflects the first sighting, with the rssi/timing block
// summarising subsequent ones.
func (a *Aggregator) Observe(key string, rssi int, recordFn func() interface{}) {
	now := time.Now()

	if a.window <= 0 {
		// Immediate mode — no shared state, no lock needed.
		a.flush(&Bucket{
			Key:       key,
			FirstSeen: now,
			LastSeen:  now,
			RSSI:      []int{rssi},
			Record:    recordFn(),
		})
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	b, ok := a.buckets[key]
	if !ok {
		a.buckets[key] = &Bucket{
			Key:       key,
			FirstSeen: now,
			LastSeen:  now,
			RSSI:      []int{rssi},
			Record:    recordFn(),
		}
		return
	}
	b.LastSeen = now
	b.RSSI = append(b.RSSI, rssi)
}

// ObserveExisting records an RSSI sample only if a bucket for key already
// exists in the current window. It returns true when the sample was merged
// into an existing bucket — letting the caller skip re-parsing a packet that
// will be discarded anyway — and false when no bucket exists yet, in which
// case the caller must parse the payload and call Observe to create it.
//
// In immediate mode (window <= 0) there are no persistent buckets, so this
// always returns false and every packet is parsed and emitted.
func (a *Aggregator) ObserveExisting(key string, rssi int) bool {
	if a.window <= 0 {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	b, ok := a.buckets[key]
	if !ok {
		return false
	}
	b.LastSeen = time.Now()
	b.RSSI = append(b.RSSI, rssi)
	return true
}

// Close stops the flush loop and drains all in-flight buckets.
func (a *Aggregator) Close() {
	close(a.stop)
	<-a.done
}

func (a *Aggregator) run() {
	defer close(a.done)
	if a.window <= 0 {
		<-a.stop
		return
	}
	tick := a.window / 4
	if tick < time.Second {
		tick = time.Second
	}
	t := time.NewTicker(tick)
	defer t.Stop()
	for {
		select {
		case <-a.stop:
			a.flushAll()
			return
		case now := <-t.C:
			a.flushExpired(now)
		}
	}
}

func (a *Aggregator) flushExpired(now time.Time) {
	a.mu.Lock()
	var ready []*Bucket
	for k, b := range a.buckets {
		if now.Sub(b.FirstSeen) >= a.window {
			ready = append(ready, b)
			delete(a.buckets, k)
		}
	}
	a.mu.Unlock()
	// Flush outside the lock so a slow sink can't block Observe.
	for _, b := range ready {
		a.flush(b)
	}
}

func (a *Aggregator) flushAll() {
	a.mu.Lock()
	all := make([]*Bucket, 0, len(a.buckets))
	for k, b := range a.buckets {
		all = append(all, b)
		delete(a.buckets, k)
	}
	a.mu.Unlock()
	// Stable order on shutdown helps tests and aids any operator tail-ing
	// the file at the moment of close.
	sort.Slice(all, func(i, j int) bool { return all[i].FirstSeen.Before(all[j].FirstSeen) })
	for _, b := range all {
		a.flush(b)
	}
}
