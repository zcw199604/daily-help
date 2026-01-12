package wecom

import (
	"sync"
	"testing"
	"time"
)

func TestDeduper_SeenOrMark_TTLLogic_NoSleep(t *testing.T) {
	t.Parallel()

	d := NewDeduper(10 * time.Minute)
	t.Cleanup(d.Close)

	if got := d.SeenOrMark("k"); got {
		t.Fatalf("SeenOrMark(1) = true, want false")
	}
	if got := d.SeenOrMark("k"); !got {
		t.Fatalf("SeenOrMark(2) = false, want true")
	}

	d.mu.Lock()
	d.data["k"] = time.Now().Add(-1 * time.Second)
	d.mu.Unlock()
	if got := d.SeenOrMark("k"); got {
		t.Fatalf("SeenOrMark(after ttl) = true, want false")
	}
}

func TestDeduper_SeenOrMark_EmptyKey(t *testing.T) {
	t.Parallel()

	d := NewDeduper(10 * time.Minute)
	t.Cleanup(d.Close)
	if got := d.SeenOrMark(""); got {
		t.Fatalf("SeenOrMark(\"\") = true, want false")
	}
}

func TestDeduper_SeenOrMark_NilReceiver(t *testing.T) {
	t.Parallel()

	var d *Deduper
	if got := d.SeenOrMark("k"); got {
		t.Fatalf("(*Deduper)(nil).SeenOrMark() = true, want false")
	}
}

func TestDeduper_NewDeduper_DefaultTTL(t *testing.T) {
	t.Parallel()

	d := NewDeduper(0)
	t.Cleanup(d.Close)
	if d.ttl != 10*time.Minute {
		t.Fatalf("ttl = %s, want %s", d.ttl, 10*time.Minute)
	}
}

func TestDeduper_PruneExpired(t *testing.T) {
	t.Parallel()

	d := NewDeduper(10 * time.Minute)
	t.Cleanup(d.Close)
	_ = d.SeenOrMark("k")

	d.mu.Lock()
	d.data["k"] = time.Now().Add(-1 * time.Second)
	d.mu.Unlock()

	d.pruneExpired()

	d.mu.Lock()
	_, ok := d.data["k"]
	d.mu.Unlock()
	if ok {
		t.Fatalf("pruneExpired did not remove expired key")
	}
}

func TestDeduper_SeenOrMark_ConcurrentSameKey(t *testing.T) {
	t.Parallel()

	d := NewDeduper(10 * time.Minute)
	t.Cleanup(d.Close)

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)

	var falseCount int
	var mu sync.Mutex
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			got := d.SeenOrMark("k")
			if !got {
				mu.Lock()
				falseCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if falseCount != 1 {
		t.Fatalf("falseCount = %d, want 1", falseCount)
	}
}
