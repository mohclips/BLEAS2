package adv

import (
	"testing"
	"time"
)

// runWithTimeout runs fn and fails the test if it does not return within d.
// A regression in the AD parser manifests as an infinite loop, which would
// otherwise hang the whole test binary — this turns that into a clear failure.
func runWithTimeout(t *testing.T, d time.Duration, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		fn()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatal("UUIDs() did not return — infinite loop in malformed AD parsing")
	}
}

// TestUUIDs_MalformedFieldDoesNotHang reproduces the busy-loop where a device
// broadcasts an AD structure whose length byte is 0 or overruns the buffer.
// Before the fix, getUUIDsByType re-parsed the same byte forever, pinning a CPU
// core per malformed advertisement.
func TestUUIDs_MalformedFieldDoesNotHang(t *testing.T) {
	cases := map[string][]byte{
		"zero-length field":  {0x00, 0x03},             // l=0
		"overrun length":     {0x05, 0x03, 0xaa},       // l=5, only 3 bytes present
		"zero after valid":   {0x03, 0x03, 0xaa, 0xbb, 0x00, 0x03}, // good UUID, then l=0
		"single trailing byte": {0x03, 0x03, 0xaa, 0xbb, 0x00},     // good UUID, then stray byte
	}
	for name, raw := range cases {
		raw := raw
		t.Run(name, func(t *testing.T) {
			p := NewRawPacket(raw)
			runWithTimeout(t, 2*time.Second, func() {
				_ = p.UUIDs()
			})
		})
	}
}

// TestUUIDs_WellFormedStillParses guards against the fix over-reaching and
// dropping valid UUIDs.
func TestUUIDs_WellFormedStillParses(t *testing.T) {
	// One complete 16-bit UUID list: len=3, type=allUUID16, data=0xaa 0xbb.
	p := NewRawPacket([]byte{0x03, allUUID16, 0xaa, 0xbb})
	var got []string
	runWithTimeout(t, 2*time.Second, func() {
		for _, u := range p.UUIDs() {
			got = append(got, u.String())
		}
	})
	if len(got) != 1 {
		t.Fatalf("want 1 UUID, got %d (%v)", len(got), got)
	}
	// UUID16 is little-endian: bytes aa bb -> 0xbbaa.
	if got[0] != "bbaa" {
		t.Errorf("want UUID bbaa, got %s", got[0])
	}
}
