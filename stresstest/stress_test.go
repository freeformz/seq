// Package stresstest contains stress and regression tests for github.com/freeformz/seq that don't fit the
// Example-based test style used in the main package: panics, hangs, data races, and goroutine leaks. Run with the
// race detector enabled (go test -race ./...).
package stresstest

import (
	"context"
	"iter"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/freeformz/seq"
)

// mustPanic fails the test if fn does not panic.
func mustPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: expected panic, got none", name)
		}
	}()
	fn()
}

// withTimeout fails the test if fn does not return within d. Guards against regressions that hang forever (e.g.
// ranging over a nil channel). On timeout the fn goroutine is abandoned, so fn must not use t.
func withTimeout(t *testing.T, d time.Duration, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		t.Fatal("timed out")
	}
}

func TestChunkPanicsOnNonPositiveSize(t *testing.T) {
	// Regression: Chunk with size < 1 used to silently accumulate the entire sequence into a single chunk.
	mustPanic(t, "Chunk size 0", func() { seq.Chunk(seq.With(1, 2, 3), 0) })
	mustPanic(t, "Chunk size -1", func() { seq.Chunk(seq.With(1, 2, 3), -1) })
}

func TestChunkKVPanicsOnNonPositiveSize(t *testing.T) {
	type kv = seq.KV[string, int]
	mustPanic(t, "ChunkKV size 0", func() { seq.ChunkKV(seq.WithKV(kv{K: "a", V: 1}), 0) })
	mustPanic(t, "ChunkKV size -1", func() { seq.ChunkKV(seq.WithKV(kv{K: "a", V: 1}), -1) })
}

func TestEveryUntilPanicsOnNonPositiveDuration(t *testing.T) {
	// Regression: time.Tick returns a nil channel for d <= 0, so iterating used to block forever instead of
	// panicking as documented.
	mustPanic(t, "EveryUntil d=0", func() { seq.EveryUntil(0, time.Now()) })
	mustPanic(t, "EveryUntil d=-1", func() { seq.EveryUntil(-time.Second, time.Now()) })
}

func TestEveryUntilTickCount(t *testing.T) {
	// On the synctest fake clock ticks land exactly on every interval, so the count is exact: ticks at 10, 20, 30,
	// and 40ms; the 50ms tick is past the deadline.
	synctest.Test(t, func(t *testing.T) {
		var ticks int
		for range seq.EveryUntil(10*time.Millisecond, time.Now().Add(45*time.Millisecond)) {
			ticks++
		}
		if ticks != 4 {
			t.Errorf("EveryUntil yielded %d ticks, want 4", ticks)
		}
	})
}

func TestEveryUntilSlowIterateeEndsWithoutExtraTick(t *testing.T) {
	// Regression: EveryUntil used to re-check the pre-yield timestamp after the yield returned, so a slow iteratee
	// that consumed the remaining time still waited out one more tick (40ms total here) before noticing the
	// deadline had passed. It must end as soon as the yield returns (at 30ms: first tick at 10ms + 20ms sleep).
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		var ticks int
		for range seq.EveryUntil(10*time.Millisecond, start.Add(15*time.Millisecond)) {
			ticks++
			time.Sleep(20 * time.Millisecond)
		}
		if ticks != 1 {
			t.Errorf("EveryUntil yielded %d ticks, want 1", ticks)
		}
		if elapsed := time.Since(start); elapsed != 30*time.Millisecond {
			t.Errorf("EveryUntil ended after %v, want 30ms", elapsed)
		}
	})
}

func TestEveryNTiming(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		var ticks int
		for range seq.EveryN(10*time.Millisecond, 3) {
			ticks++
		}
		if ticks != 3 {
			t.Errorf("EveryN yielded %d ticks, want 3", ticks)
		}
		if elapsed := time.Since(start); elapsed != 30*time.Millisecond {
			t.Errorf("EveryN ended after %v, want 30ms", elapsed)
		}
	})
}

func TestEveryNPanicsOnNonPositiveDuration(t *testing.T) {
	mustPanic(t, "EveryN d=0", func() { seq.EveryN(0, 1) })
	mustPanic(t, "EveryN d=-1", func() { seq.EveryN(-time.Second, 1) })
}

func TestEveryNNonPositiveTimesIsEmpty(t *testing.T) {
	// Regression: a negative times used to decrement past zero and tick forever.
	for _, times := range []int{0, -1, -100} {
		var yielded atomic.Bool
		withTimeout(t, 5*time.Second, func() {
			for range seq.EveryN(time.Millisecond, times) {
				yielded.Store(true)
				return
			}
		})
		if yielded.Load() {
			t.Errorf("EveryN(_, %d) yielded a value; want empty sequence", times)
		}
	}
}

func TestDropKVConcurrentIteration(t *testing.T) {
	// Regression: DropKV kept its element counter outside the iterator closure, so the returned sequence was
	// single-use and racy when iterated concurrently.
	pairs := make([]seq.KV[int, int], 100)
	for i := range pairs {
		pairs[i] = seq.KV[int, int]{K: i, V: i}
	}
	d := seq.DropKV(seq.WithKV(pairs...), 50)

	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			for range 50 {
				n := 0
				for k := range d {
					if k < 50 {
						t.Errorf("DropKV yielded dropped key %d", k)
						return
					}
					n++
				}
				if n != 50 {
					t.Errorf("DropKV iteration yielded %d pairs, want 50", n)
					return
				}
			}
		})
	}
	wg.Wait()
}

func TestIntKConcurrent(t *testing.T) {
	// IntK documents that the returned function is safe to call concurrently: hammer it and verify every value in
	// [0, goroutines*perG) is produced exactly once.
	const goroutines, perG = 32, 1000
	k := seq.IntK[struct{}]()

	results := make([][]int, goroutines)
	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Go(func() {
			out := make([]int, 0, perG)
			for range perG {
				out = append(out, k(struct{}{}))
			}
			results[g] = out
		})
	}
	wg.Wait()

	seen := make(map[int]bool, goroutines*perG)
	for _, out := range results {
		for _, v := range out {
			if v < 0 || v >= goroutines*perG {
				t.Fatalf("IntK produced out-of-range value %d", v)
			}
			if seen[v] {
				t.Fatalf("IntK produced duplicate value %d", v)
			}
			seen[v] = true
		}
	}
}

func TestCompareFuncDoesNotLeakGoroutines(t *testing.T) {
	baseline := runtime.NumGoroutine()

	for range 100 {
		// Early exits: first elements differ, a shorter, b shorter, and fully equal.
		seq.Compare(seq.With(9, 2, 3), seq.With(1, 2, 3))
		seq.Compare(seq.With(1), seq.With(1, 2, 3))
		seq.Compare(seq.With(1, 2, 3), seq.With(1))
		seq.Compare(seq.With(1, 2, 3), seq.With(1, 2, 3))
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		if runtime.NumGoroutine() <= baseline+2 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("goroutines did not drain: baseline %d, now %d", baseline, runtime.NumGoroutine())
		}
		runtime.Gosched()
		time.Sleep(10 * time.Millisecond)
	}
}

func TestToChanCtxCancelClosesChannel(t *testing.T) {
	naturals := func(yield func(int) bool) {
		for i := 0; ; i++ {
			if !yield(i) {
				return
			}
		}
	}

	ctx, cancel := context.WithCancel(t.Context())
	ch := seq.ToChanCtx(ctx, iter.Seq[int](naturals))
	for range 5 {
		<-ch
	}
	cancel()

	withTimeout(t, 5*time.Second, func() {
		for range ch { //nolint:revive // drain until closed
		}
	})
}

func TestWindowsPanicsOnNonPositiveSize(t *testing.T) {
	mustPanic(t, "Windows size 0", func() { seq.Windows(seq.With(1, 2, 3), 0) })
	mustPanic(t, "Windows size -1", func() { seq.Windows(seq.With(1, 2, 3), -1) })
}

func TestWindowsKVPanicsOnNonPositiveSize(t *testing.T) {
	type kv = seq.KV[string, int]
	mustPanic(t, "WindowsKV size 0", func() { seq.WindowsKV(seq.WithKV(kv{K: "a", V: 1}), 0) })
	mustPanic(t, "WindowsKV size -1", func() { seq.WindowsKV(seq.WithKV(kv{K: "a", V: 1}), -1) })
}

func TestCycleEmptySequenceTerminates(t *testing.T) {
	// Cycle restarts its input forever; an empty input must end the sequence instead of spinning.
	withTimeout(t, 5*time.Second, func() {
		for range seq.Cycle(seq.With[int]()) {
		}
	})
	withTimeout(t, 5*time.Second, func() {
		for range seq.CycleKV(seq.WithKV[string, int]()) {
		}
	})
}

func TestFromChanCtxCancelUnblocks(t *testing.T) {
	// FromChanCtx must end when the context is canceled even if the channel never produces another value. Inside
	// the synctest bubble a regression shows up as a deadlock panic instead of needing a real-time timeout.
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		ch := make(chan int)
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()
		for range seq.FromChanCtx(ctx, ch) {
			t.Error("FromChanCtx yielded a value from an empty channel")
		}
	})
}

func TestFromChanCtxCanceledCtxWinsOverReadyChannel(t *testing.T) {
	// Regression: with a bare two-case select, an already-canceled context raced a ready channel and values could
	// still be yielded after cancellation. Cancellation must take priority.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	ch := make(chan int, 1)
	ch <- 1
	for range 100 {
		for range seq.FromChanCtx(ctx, ch) {
			t.Fatal("FromChanCtx yielded a value after the context was already canceled")
		}
	}
}
