package seq

import (
	"cmp"
	"context"
	"iter"
	"sync/atomic"
	"time"
)

// With returns a sequence with the provided values. The values are iterated over lazily when the returned sequence is iterated
// over.
func With[T any](v ...T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, t := range v {
			if !yield(t) {
				return
			}
		}
	}
}

// KV pairs a key and a value together. Easiest way to use this is by declaring a local type with the K and V types you want
// to use and then use that, like so:
//
//	func(...) {
//		type lKV = KV[string, string]}
//		i := WithKV(lKV{"a", "1"}, lKV{"b", "2"}, lKV{"c", "3"})
//	...
type KV[K, V any] struct {
	K K
	V V
}

// WithKV returns a sequence with the provided key-value pairs. The key-value pairs are iterated over lazily when the returned
// sequence is iterated over.
func WithKV[K, V any](kv ...KV[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, kv := range kv {
			if !yield(kv.K, kv.V) {
				return
			}
		}
	}
}

// FromChan returns a sequence that yields values from the provided channel. The sequence is iterated over lazily when the
// returned sequence is iterated over. The sequence will end when the channel is closed.
//
// This allows for collecting values from a channel into a slice or similar relatively easily:
//
//	s := slices.Collect(FromChan(ch))
//	// instead of
//	var s []T
//	for v := range ch {
//		s = append(s, v)
//	}
func FromChan[T any](ch <-chan T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for t := range ch {
			if !yield(t) {
				return
			}
		}
	}
}

// ToChan returns a channel that yields values from the provided sequence. The provided sequence is iterated over lazily when
// the returned channel is iterated over. The channel is closed when the sequence is exhausted. If the consumer stops
// receiving before the sequence is exhausted, the producing goroutine blocks forever; use [ToChanCtx] when the
// consumer may abandon the channel.
func ToChan[T any](seq iter.Seq[T]) <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for t := range seq {
			ch <- t
		}
	}()
	return ch
}

// ToChanCtx returns a channel that yields values from the provided sequence. The provided sequence is iterated over
// lazily when the returned channel is iterated over. The channel is closed when the sequence is exhausted or the
// context is canceled, whichever comes first.
func ToChanCtx[T any](ctx context.Context, seq iter.Seq[T]) <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for t := range seq {
			select {
			case <-ctx.Done():
				return
			default:
			}
			select {
			case <-ctx.Done():
				return
			case ch <- t:
			}
		}
	}()
	return ch
}

// Map the values in the sequence to a new sequence of values by applying the function fn to each value. Function application
// happens lazily when the returned sequence is iterated over.
func Map[T, O any](seq iter.Seq[T], fn func(T) O) iter.Seq[O] {
	return func(yield func(o O) bool) {
		for o := range seq {
			if !yield(fn(o)) {
				return
			}
		}
	}
}

// MapKV maps the key-value pairs in the sequence to a new sequence of key-value pairs by applying the function fn to each
// key-value pair. Function application happens lazily when the returned sequence is iterated over.
func MapKV[K, V, K1, V1 any](seq iter.Seq2[K, V], fn func(K, V) (K1, V1)) iter.Seq2[K1, V1] {
	return func(yield func(K1, V1) bool) {
		for k, v := range seq {
			if !yield(fn(k, v)) {
				return
			}
		}
	}
}

// Append the items to the sequence and return an extended sequence. The provided sequence and appended items are iterated over
// lazily when the returned sequence is iterated over.
func Append[T any](seq iter.Seq[T], items ...T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for item := range seq {
			if !yield(item) {
				return
			}
		}
		for _, item := range items {
			if !yield(item) {
				return
			}
		}
	}
}

// AppendKV appends the key-value pairs to the sequence and returns an extended sequence. The provided sequence and appended
// key-value pairs are iterated over lazily when the returned sequence is iterated over.
func AppendKV[K, V any](seq iter.Seq2[K, V], items ...KV[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range seq {
			if !yield(k, v) {
				return
			}
		}
		for _, kv := range items {
			if !yield(kv.K, kv.V) {
				return
			}
		}
	}
}

// Filter the values in the sequence by applying fn to each value. Filtering happens when the returned sequence is
// iterated over.
func Filter[T any](seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		seq(func(t T) bool {
			if fn(t) {
				return yield(t)
			}
			return true
		})
	}
}

// FilterKV filters the key-value pairs in the sequence by applying fn to each key-value pair. Filtering happens when the
// returned sequence is iterated over.
func FilterKV[K, V any](seq iter.Seq2[K, V], fn func(K, V) bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range seq {
			if fn(k, v) {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// IntK returns a function that returns an increasing integer each time it is called, starting at 0. The returned function is stateful
// and is safe to call concurrently.
func IntK[V any]() func(V) int {
	var i atomic.Int64
	i.Store(-1)
	return func(V) int {
		return int(i.Add(1))
	}
}

// IterKV converts an iter.Seq[V] to an iter.Seq2[K, V]. The provided sequence is iterated over lazily when the returned
// sequence is iterated over. keyFn is called for each value to get the key.
func IterKV[K, V any](seq iter.Seq[V], keyFn func(V) K) iter.Seq2[K, V] {
	return func(yield func(k K, v V) bool) {
		for v := range seq {
			k := keyFn(v)
			if !yield(k, v) {
				return
			}
		}
	}
}

// IterK converts an iter.Seq2[K, V] to an iter.Seq[K]. The provided sequence is iterated over lazily when the returned
// sequence is iterated over.
func IterK[K, V any](seq iter.Seq2[K, V]) iter.Seq[K] {
	return func(yield func(k K) bool) {
		for k := range seq {
			if !yield(k) {
				return
			}
		}
	}
}

// IterV converts an iter.Seq2[K, V] to an iter.Seq[V]. The provided sequence is iterated over lazily when the returned
// sequence is iterated over.
func IterV[K, V any](seq iter.Seq2[K, V]) iter.Seq[V] {
	return func(yield func(v V) bool) {
		for _, v := range seq {
			if !yield(v) {
				return
			}
		}
	}
}

// Max value of the sequence. Uses max builtin to compare values. The second value is false if the sequence is empty. The
// sequence is iterated over before Max returns.
func Max[T cmp.Ordered](seq iter.Seq[T]) (T, bool) {
	var mt T
	var value bool
	for t := range seq {
		if !value {
			mt = t
			value = true
		} else {
			mt = max(t, mt)
		}
	}
	return mt, value
}

// MaxFunc is like [Max] but uses the function to compare elements. The provided sequence is iterated over before MaxFunc returns.
func MaxFunc[T any](seq iter.Seq[T], compare func(T, T) int) (T, bool) {
	var mt T
	var value bool
	for t := range seq {
		if !value {
			mt = t
			value = true
		} else if compare(t, mt) > 0 {
			mt = t
		}
	}
	return mt, value
}

// MaxFuncKV is like [MaxFunc] but for key-value pairs. The provided sequence is iterated over before MaxFuncKV returns.
func MaxFuncKV[K, V any](seq iter.Seq2[K, V], compare func(KV[K, V], KV[K, V]) int) (KV[K, V], bool) {
	var mt KV[K, V]
	var value bool
	for k, v := range seq {
		t := KV[K, V]{K: k, V: v}
		if !value {
			mt = t
			value = true
		} else if compare(t, mt) > 0 {
			mt = t
		}
	}
	return mt, value
}

// Min value from the sequence. Uses min built in to compare values. The second value is false if the sequence is empty. The
// sequence is iterated over before Min returns.
func Min[T cmp.Ordered](seq iter.Seq[T]) (T, bool) {
	var mt T
	var value bool
	for t := range seq {
		if !value {
			mt = t
			value = true
		} else {
			mt = min(t, mt)
		}
	}
	return mt, value
}

// MinFunc is like [Min] but uses the function to compare elements. The provided sequence is iterated over before MinFunc returns.
func MinFunc[T any](seq iter.Seq[T], compare func(T, T) int) (T, bool) {
	var mt T
	var value bool
	for t := range seq {
		if !value {
			mt = t
			value = true
		} else if compare(t, mt) < 0 {
			mt = t
		}
	}
	return mt, value
}

// MinFuncKV is like [MinFunc] but for key-value pairs. The provided sequence is iterated over before MinFuncKV returns.
func MinFuncKV[K, V any](seq iter.Seq2[K, V], compare func(KV[K, V], KV[K, V]) int) (KV[K, V], bool) {
	var mt KV[K, V]
	var value bool
	for k, v := range seq {
		t := KV[K, V]{K: k, V: v}
		if !value {
			mt = t
			value = true
		} else if compare(t, mt) < 0 {
			mt = t
		}
	}
	return mt, value
}

// Reduce the sequence to a single value by applying the function fn to each value. The provided sequence is iterated
// over before Reduce returns.
func Reduce[T, O any](seq iter.Seq[T], initial O, fn func(agg O, t T) O) O {
	agg := initial
	for t := range seq {
		agg = fn(agg, t)
	}
	return agg
}

// ReduceKV reduces the sequence to a single value by applying the function fn to each key-value pair. The provided sequence is iterated
// over before ReduceKV returns.
func ReduceKV[K, V, O any](seq iter.Seq2[K, V], initial O, fn func(agg O, k K, v V) O) O {
	agg := initial
	for k, v := range seq {
		agg = fn(agg, k, v)
	}
	return agg
}

// Compact returns an iterator that yields all values that are not equal to the previous value. The provided sequence is iterated
// over lazily when the returned sequence is iterated over.
func Compact[T comparable](seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		var prev T
		first := true
		for t := range seq {
			if first || prev != t {
				first = false
				prev = t
				if !yield(t) {
					return
				}
			}
		}
	}
}

// CompactFunc is like [Compact] but uses an the function to compare elements. For runs of elements that compare equal,
// CompactFunc only yields the first one. The provided sequence is iterated over lazily when the returned sequence is
// iterated over.
func CompactFunc[T any](seq iter.Seq[T], equal func(T, T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		var prev T
		first := true
		for t := range seq {
			if first || !equal(prev, t) {
				first = false
				prev = t
				if !yield(t) {
					return
				}
			}
		}
	}
}

// CompactKV returns an iterator that yields all key-value pairs that are not equal to the previous key-value pair. The provided
// sequence is iterated over lazily when the returned sequence is iterated over.
func CompactKV[K, V comparable](seq iter.Seq2[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		var prev KV[K, V]
		first := true
		for k, v := range seq {
			if first || prev.K != k || prev.V != v {
				first = false
				prev.K = k
				prev.V = v
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// CompactKVFunc is like [CompactKV] but uses the function to compare key-value pairs. For runs of key-value pairs that compare
// equal, CompactKVFunc only yields the first one. The provided sequence is iterated over lazily when the returned sequence is
// iterated over.
func CompactKVFunc[K, V any](seq iter.Seq2[K, V], equal func(KV[K, V], KV[K, V]) bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		var prev KV[K, V]
		first := true
		for k, v := range seq {
			if first || !equal(prev, KV[K, V]{K: k, V: v}) {
				first = false
				prev.K = k
				prev.V = v
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// Chunk the sequence into chunks of size. The provided sequence is iterated over lazily when the returned sequence is iterated
// over. The last chunk may have fewer than size elements. The size must be at least 1; if not, the function will panic.
func Chunk[T any](seq iter.Seq[T], size int) iter.Seq[iter.Seq[T]] {
	if size < 1 {
		panic("seq: Chunk size must be at least 1")
	}
	return func(yield func(iter.Seq[T]) bool) {
		var chunk []T
		// The first chunk grows via append so a sequence shorter than size never
		// over-allocates; once a chunk has filled, later ones preallocate exactly size.
		full := false
		for t := range seq {
			if chunk == nil && full {
				chunk = make([]T, 0, size)
			}
			chunk = append(chunk, t)
			if len(chunk) == size {
				full = true
				if !yield(With(chunk...)) {
					return
				}
				chunk = nil
			}
		}
		if len(chunk) > 0 {
			yield(With(chunk...))
		}
	}
}

// ChunkKV is like [Chunk] but for key-value pairs. The provided sequence is iterated over lazily when the returned sequence is
// iterated over. The last chunk may have fewer than size elements. The size must be at least 1; if not, the function will panic.
func ChunkKV[K, V any](seq iter.Seq2[K, V], size int) iter.Seq[iter.Seq2[K, V]] {
	if size < 1 {
		panic("seq: ChunkKV size must be at least 1")
	}
	return func(yield func(iter.Seq2[K, V]) bool) {
		var chunk []KV[K, V]
		// The first chunk grows via append so a sequence shorter than size never
		// over-allocates; once a chunk has filled, later ones preallocate exactly size.
		full := false
		for k, v := range seq {
			if chunk == nil && full {
				chunk = make([]KV[K, V], 0, size)
			}
			chunk = append(chunk, KV[K, V]{K: k, V: v})
			if len(chunk) == size {
				full = true
				if !yield(WithKV(chunk...)) {
					return
				}
				chunk = nil
			}
		}
		if len(chunk) > 0 {
			yield(WithKV(chunk...))
		}
	}
}

// Compare is like [CompareFunc] but uses the cmp.Compare function to compare elements.
func Compare[T cmp.Ordered](a, b iter.Seq[T]) int {
	return CompareFunc(a, b, cmp.Compare)
}

// CompareFunc compares the elements of a and b, using the compare func on each pair of elements. The elements are
// compared sequentially, until one element is not equal to the other. The result of comparing the first non-matching
// elements is returned verbatim — it is not normalized to -1 or +1. If both sequences are equal until one of them
// ends, the shorter sequence is considered less than the longer one: the result is +1 if a is longer, -1 if b is
// longer, and 0 if the sequences are equal.
func CompareFunc[T any](a, b iter.Seq[T], compare func(T, T) int) int {
	next, stop := iter.Pull(b)
	defer stop()

	for av := range a {
		bv, ok := next()
		if !ok { // b is shorter than a
			return 1
		}
		if c := compare(av, bv); c != 0 {
			return c
		}
	}

	// done with a, check if b is longer
	if _, ok := next(); ok {
		return -1
	}

	// a and b are equal
	return 0
}

// CompareKV is like [CompareKVFunc] but uses the cmp.Compare function to compare keys and values.
func CompareKV[K, V cmp.Ordered](a, b iter.Seq2[K, V]) int {
	return CompareKVFunc(a, b, func(a, b KV[K, V]) int {
		if c := cmp.Compare(a.K, b.K); c != 0 {
			return c
		}
		return cmp.Compare(a.V, b.V)
	})
}

// CompareKVFunc compares the key-value pairs of a and b, using the compare func on each pair of key-value pairs. The key-value
// pairs are compared sequentially, until one key-value pair is not equal to the other. The result of comparing the first
// non-matching key-value pairs is returned verbatim — it is not normalized to -1 or +1. If both sequences are equal
// until one of them ends, the shorter sequence is considered less than the longer one: the result is +1 if a is
// longer, -1 if b is longer, and 0 if the sequences are equal.
func CompareKVFunc[AK, AV, BK, BV any](a iter.Seq2[AK, AV], b iter.Seq2[BK, BV], compare func(a KV[AK, AV], b KV[BK, BV]) int) int {
	next, stop := iter.Pull2(b)
	defer stop()

	for ak, av := range a {
		bk, bv, ok := next()
		if !ok { // b is shorter than a
			return 1
		}
		if c := compare(KV[AK, AV]{K: ak, V: av}, KV[BK, BV]{K: bk, V: bv}); c != 0 {
			return c
		}
	}

	// done with a, check if b is longer
	if _, _, ok := next(); ok {
		return -1
	}

	// a and b are equal
	return 0
}

// Contains returns true if the value is in the sequence. The sequence is iterated over when Contains is called.
func Contains[T comparable](seq iter.Seq[T], value T) bool {
	for t := range seq {
		if t == value {
			return true
		}
	}
	return false
}

// ContainsKV returns true if the key-value pair is in the sequence. The sequence is iterated over when ContainsKV is called.
func ContainsKV[K, V comparable](seq iter.Seq2[K, V], key K, value V) bool {
	for k, v := range seq {
		if k == key && v == value {
			return true
		}
	}
	return false
}

// ContainsFunc returns true if the predicate function returns true for any value in the sequence. The sequence is
// iterated over when ContainsFunc is called.
func ContainsFunc[T any](seq iter.Seq[T], predicate func(T) bool) bool {
	for t := range seq {
		if predicate(t) {
			return true
		}
	}
	return false
}

// ContainsKVFunc returns true if the predicate function returns true for any key-value pair in the sequence. The
// sequence is iterated over when ContainsKVFunc is called.
func ContainsKVFunc[K, V any](seq iter.Seq2[K, V], predicate func(K, V) bool) bool {
	for k, v := range seq {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// Equal returns true if the sequences are equal. The sequences are compared sequentially, until one element is not equal to
// the other.
func Equal[T comparable](a, b iter.Seq[T]) bool {
	return CompareFunc(a, b, func(a, b T) int {
		if a == b {
			return 0
		}
		return -1
	}) == 0
}

// EqualKV returns true if the key-value pairs in the sequences are equal. The key-value pairs are compared sequentially, until
// one key-value pair is not equal to the other.
func EqualKV[K, V comparable](a, b iter.Seq2[K, V]) bool {
	return CompareKVFunc(a, b, func(a, b KV[K, V]) int {
		if a.K == b.K && a.V == b.V {
			return 0
		}
		if a.K != b.K {
			return -1
		}
		return 1
	}) == 0
}

// EqualFunc is like [Equal] but uses the function to compare elements.
func EqualFunc[T any](a, b iter.Seq[T], equal func(T, T) bool) bool {
	return CompareFunc(a, b, func(a, b T) int {
		if equal(a, b) {
			return 0
		}
		return -1
	}) == 0
}

// EqualKVFunc is like [EqualKV] but uses the function to compare key-value pairs.
func EqualKVFunc[AK, AV, BK, BV any](a iter.Seq2[AK, AV], b iter.Seq2[BK, BV], equal func(a KV[AK, AV], b KV[BK, BV]) bool) bool {
	return CompareKVFunc(a, b, func(a KV[AK, AV], b KV[BK, BV]) int {
		if equal(a, b) {
			return 0
		}
		return 1
	}) == 0
}

// Repeat returns a sequence which repeats the value n times.
func Repeat[T any](n int, t T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for range n {
			if !yield(t) {
				return
			}
		}
	}
}

// RepeatKV returns a sequence which repeats the key-value pair n times.
func RepeatKV[K, V any](n int, k K, v V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for range n {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Replace the old value with the new value in the sequence. The provided sequence is iterated over lazily when the
// returned sequence is iterated over.
func Replace[T comparable](seq iter.Seq[T], old, new T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for t := range seq {
			if t == old {
				t = new
			}
			if !yield(t) {
				return
			}
		}
	}
}

// ReplaceKV replaces the old key-value pair with the new key-value pair in the sequence. The provided sequence is iterated
// over lazily when the returned sequence is iterated over.
func ReplaceKV[K, V comparable](seq iter.Seq2[K, V], old KV[K, V], new KV[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range seq {
			if k == old.K && v == old.V {
				k = new.K
				v = new.V
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// IsSorted returns true if the sequence is sorted. The provided sequence is iterated over before IsSorted returns.
// [cmp.Compare] is used to compare elements.
func IsSorted[T cmp.Ordered](seq iter.Seq[T]) bool {
	var prev T
	first := true
	for t := range seq {
		if !first && cmp.Compare(t, prev) < 0 {
			return false
		}
		first = false
		prev = t
	}
	return true
}

// IsSortedKV returns true if the sequence is sorted. The keys and values are each compared independently with
// [cmp.Compare]: both the keys and the values must be non-decreasing. Note this is stricter than the lexicographic
// (key, then value) ordering used by [CompareKV]. The provided sequence is iterated over before IsSortedKV returns.
func IsSortedKV[K, V cmp.Ordered](seq iter.Seq2[K, V]) bool {
	var prev KV[K, V]
	first := true
	for k, v := range seq {
		if !first && ((cmp.Compare(k, prev.K) < 0) || (cmp.Compare(v, prev.V) < 0)) {
			return false
		}
		first = false
		prev.K = k
		prev.V = v
	}
	return true
}

// Coalesce returns the first non zero value in the sequence. The provided sequence is iterated over when Coalesce is
// called, stopping at the first non-zero value. If no non-zero value is found, the second return value is false.
func Coalesce[T comparable](seq iter.Seq[T]) (T, bool) {
	var zero T
	for t := range seq {
		if t != zero {
			return t, true
		}
	}
	return zero, false
}

// CoalesceKV returns the first key-value pair in the sequence whose value is non zero. The provided sequence is
// iterated over when CoalesceKV is called, stopping at the first non-zero value. If no non-zero value is found, the
// second return value is false.
func CoalesceKV[K, V comparable](seq iter.Seq2[K, V]) (KV[K, V], bool) {
	var zero V
	for k, v := range seq {
		if v != zero {
			return KV[K, V]{K: k, V: v}, true
		}
	}
	return KV[K, V]{}, false
}

// Count returns the number of elements in the sequence. The sequence is iterated over before Count returns.
func Count[T any](seq iter.Seq[T]) int {
	var count int
	for range seq {
		count++
	}
	return count
}

// CountKV returns the number of key-value pairs in the sequence. The sequence is iterated over before CountKV returns.
func CountKV[K, V any](seq iter.Seq2[K, V]) int {
	var count int
	for range seq {
		count++
	}
	return count
}

// CountBy returns the number of elements in the sequence for which the function returns true. The sequence is iterated over
// before CountBy returns.
func CountBy[T any](seq iter.Seq[T], fn func(T) bool) int {
	var count int
	for t := range seq {
		if fn(t) {
			count++
		}
	}
	return count
}

// CountKVBy returns the number of key-value pairs in the sequence for which the function returns true. The sequence is
// iterated over before CountKVBy returns.
func CountKVBy[K, V any](seq iter.Seq2[K, V], fn func(K, V) bool) int {
	var count int
	for k, v := range seq {
		if fn(k, v) {
			count++
		}
	}
	return count
}

// CountValues returns a key-value sequence where the keys are the values in the original sequence and the values are
// the number of times that value appears in the original sequence. The returned key-value sequence is unordered. The
// provided sequence is iterated over before CountValues returns.
func CountValues[T comparable](seq iter.Seq[T]) iter.Seq2[T, int] {
	m := make(map[T]int)
	for t := range seq {
		m[t]++
	}
	return func(yield func(T, int) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Drop n elements from the starts of the sequence. The provided sequence is iterated over lazily when the returned
// sequence is iterated over.
func Drop[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		i := -1
		for t := range seq {
			i++
			if i < n {
				continue
			}
			if !yield(t) {
				return
			}
		}
	}
}

// DropKV n key-value pairs from the starts of the sequence. The provided sequence is iterated over lazily when the returned
// sequence is iterated over.
func DropKV[K, V any](seq iter.Seq2[K, V], n int) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		i := -1
		for k, v := range seq {
			i++
			if i < n {
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// DropBy returns a sequence with all elements for which the function returns true removed. The provided sequence is
// iterated over lazily when the returned sequence is iterated over. This is the opposite of Filter.
func DropBy[T any](seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
	return Filter(seq, func(t T) bool {
		return !fn(t)
	})
}

// DropKVBy returns a sequence with all key-value pairs for which the function returns true removed. The provided sequence
// is iterated over lazily when the returned sequence is iterated over. This is the opposite of FilterKV.
func DropKVBy[K, V any](seq iter.Seq2[K, V], fn func(K, V) bool) iter.Seq2[K, V] {
	return FilterKV(seq, func(k K, v V) bool {
		return !fn(k, v)
	})
}

// EveryUntil returns a sequence that yields the time every d duration until the provided time. The ticker will adjust
// the time interval or drop ticks to make up for slow iteratee. The duration d must be greater than zero; if not,
// the function will panic. Waits d long before yielding the first element.
func EveryUntil(d time.Duration, until time.Time) iter.Seq[time.Time] {
	if d <= 0 {
		panic("seq: EveryUntil interval must be positive")
	}
	return func(yield func(time.Time) bool) {
		for now := range time.Tick(d) {
			if now.After(until) {
				return
			}
			if !yield(now) {
				return
			}
			// Re-check the clock after the yield returns: a slow iteratee may have consumed the
			// remaining time, and ending here beats waiting out another tick to notice. Checking
			// now again would be useless — it cannot have changed since the check above.
			if time.Now().After(until) {
				return
			}
		}
	}
}

// EveryN returns a sequence that yields the time every d duration n times. The ticker will adjust the time interval or
// drop ticks to make up for slow iteratee. The duration d must be greater than zero; if not, the function will panic.
// Waits d long before yielding the first element. If times is not positive, the sequence is empty.
func EveryN(d time.Duration, times int) iter.Seq[time.Time] {
	if d <= 0 {
		panic("seq: EveryN interval must be positive")
	}
	return func(yield func(time.Time) bool) {
		if times <= 0 {
			return
		}
		for now := range time.Tick(d) {
			if !yield(now) {
				return
			}
			times--
			if times == 0 {
				return
			}
		}
	}
}

// MapToKV maps the values in the sequence to a new sequence of key-value pairs by applying the function fn to each value. Function
// application happens lazily when the returned sequence is iterated over.
func MapToKV[T, K, V any](seq iter.Seq[T], fn func(T) (K, V)) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for t := range seq {
			k, v := fn(t)
			if !yield(k, v) {
				return
			}
		}
	}
}

// At returns the value at the given 0-based index in the sequence and true. If
// the index is out of range (negative or beyond sequence length), it returns
// the zero value and false. The provided sequence is iterated over up to and
// including the target index when At is called.
func At[T any](seq iter.Seq[T], index int) (T, bool) {
	if index < 0 {
		var z T
		return z, false
	}
	var i int
	for v := range seq {
		if i == index {
			return v, true
		}
		i++
	}
	var z T
	return z, false
}

// AtKV returns the key and value at the given 0-based index in the sequence and true. If
// the index is out of range (negative or beyond sequence length), it returns
// the zero values and false. The provided sequence is iterated over up to and
// including the target index when AtKV is called. To find a value by key, use
// FindByKey instead.
func AtKV[K any, V any](seq iter.Seq2[K, V], index int) (K, V, bool) {
	if index < 0 {
		var zk K
		var zv V
		return zk, zv, false
	}
	var i int
	for k, v := range seq {
		if i == index {
			return k, v, true
		}
		i++
	}
	var zk K
	var zv V
	return zk, zv, false
}

// Find returns the 0-based index of the first occurrence of the value in the sequence and true. If the value is not
// found, the first return value is the length of the sequence and the second return value is false. The provided
// sequence is iterated over when Find is called.
func Find[T comparable](seq iter.Seq[T], value T) (int, bool) {
	var i int
	for t := range seq {
		if t == value {
			return i, true
		}
		i++
	}
	return i, false
}

// FindBy returns the first value in the sequence for which the function returns true, the "index" (0 based) of the
// value, and true. If no value is found, the first return value is the zero value of the type, the second return value
// is the length of the sequence, and the third return value is false. The provided sequence is iterated over when FindBy is called.
func FindBy[T any](seq iter.Seq[T], fn func(T) bool) (T, int, bool) {
	var i int
	for t := range seq {
		if fn(t) {
			return t, i, true
		}
		i++
	}
	var z T
	return z, i, false
}

// FindByKey returns the value of the first key-value pair in the sequence for which the function returns true, the
// "index" (0 based) of the value, and true. If the key is not found, the first return value is the zero value of the
// value type, the second return value is the length of the sequence, and the third return value is false. The provided
// sequence is iterated over when FindByKey is called.
func FindByKey[K comparable, V any](seq iter.Seq2[K, V], key K) (V, int, bool) {
	var i int
	for k, v := range seq {
		if k == key {
			return v, i, true
		}
		i++
	}
	var v V
	return v, i, false
}

// FindByValue is like FindByKey, but returns the key of the first key-value pair whose value is equal to the provided value.
func FindByValue[K comparable, V comparable](seq iter.Seq2[K, V], value V) (K, int, bool) {
	var i int
	for k, v := range seq {
		if v == value {
			return k, i, true
		}
		i++
	}
	var k K
	return k, i, false
}

// Take returns a sequence of the first n elements of the sequence. If the sequence has fewer than n elements, the
// returned sequence yields all of them. If n is not positive, the returned sequence is empty. The provided sequence is
// iterated over lazily when the returned sequence is iterated over.
func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		i := 0
		for t := range seq {
			if !yield(t) {
				return
			}
			i++
			if i == n {
				return
			}
		}
	}
}

// TakeKV returns a sequence of the first n key-value pairs of the sequence. If the sequence has fewer than n pairs, the
// returned sequence yields all of them. If n is not positive, the returned sequence is empty. The provided sequence is
// iterated over lazily when the returned sequence is iterated over.
func TakeKV[K, V any](seq iter.Seq2[K, V], n int) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if n <= 0 {
			return
		}
		i := 0
		for k, v := range seq {
			if !yield(k, v) {
				return
			}
			i++
			if i == n {
				return
			}
		}
	}
}

// TakeWhile returns a sequence of the leading elements of the sequence for which the function returns true. The
// sequence ends before the first element for which the function returns false. The provided sequence is iterated over
// lazily when the returned sequence is iterated over.
func TakeWhile[T any](seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for t := range seq {
			if !fn(t) || !yield(t) {
				return
			}
		}
	}
}

// TakeKVWhile returns a sequence of the leading key-value pairs of the sequence for which the function returns true.
// The sequence ends before the first pair for which the function returns false. The provided sequence is iterated over
// lazily when the returned sequence is iterated over.
func TakeKVWhile[K, V any](seq iter.Seq2[K, V], fn func(K, V) bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range seq {
			if !fn(k, v) || !yield(k, v) {
				return
			}
		}
	}
}

// DropWhile returns a sequence that skips the leading elements of the sequence for which the function returns true and
// then yields every remaining element, starting with the first element for which the function returns false. Unlike
// [DropBy], the function is not applied after the first non-matching element. The provided sequence is iterated over
// lazily when the returned sequence is iterated over.
func DropWhile[T any](seq iter.Seq[T], fn func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		dropping := true
		for t := range seq {
			if dropping && fn(t) {
				continue
			}
			dropping = false
			if !yield(t) {
				return
			}
		}
	}
}

// DropKVWhile returns a sequence that skips the leading key-value pairs of the sequence for which the function returns
// true and then yields every remaining pair, starting with the first pair for which the function returns false. Unlike
// [DropKVBy], the function is not applied after the first non-matching pair. The provided sequence is iterated over
// lazily when the returned sequence is iterated over.
func DropKVWhile[K, V any](seq iter.Seq2[K, V], fn func(K, V) bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		dropping := true
		for k, v := range seq {
			if dropping && fn(k, v) {
				continue
			}
			dropping = false
			if !yield(k, v) {
				return
			}
		}
	}
}

// Concat returns a sequence that yields the elements of each provided sequence in order. The provided sequences are
// iterated over lazily when the returned sequence is iterated over.
func Concat[T any](seqs ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, seq := range seqs {
			for t := range seq {
				if !yield(t) {
					return
				}
			}
		}
	}
}

// ConcatKV returns a sequence that yields the key-value pairs of each provided sequence in order. The provided
// sequences are iterated over lazily when the returned sequence is iterated over.
func ConcatKV[K, V any](seqs ...iter.Seq2[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, seq := range seqs {
			for k, v := range seq {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// Zip returns a sequence that pairs the elements of a and b positionally, yielding the elements of a as keys and the
// elements of b as values. The sequence ends when either input sequence ends. The provided sequences are iterated over
// lazily when the returned sequence is iterated over.
func Zip[A, B any](a iter.Seq[A], b iter.Seq[B]) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		next, stop := iter.Pull(b)
		defer stop()
		for av := range a {
			bv, ok := next()
			if !ok {
				return
			}
			if !yield(av, bv) {
				return
			}
		}
	}
}

// Merge merges two sorted sequences into one sorted sequence. [cmp.Compare] is used to compare elements. If the input
// sequences are not sorted the output will not be sorted either, but it will still contain every element of both. The
// provided sequences are iterated over lazily when the returned sequence is iterated over.
func Merge[T cmp.Ordered](a, b iter.Seq[T]) iter.Seq[T] {
	return MergeFunc(a, b, cmp.Compare)
}

// MergeFunc is like [Merge] but uses the function to compare elements. When elements compare equal, elements from b are
// yielded before elements from a. The provided sequences are iterated over lazily when the returned sequence is
// iterated over.
func MergeFunc[T any](a, b iter.Seq[T], compare func(T, T) int) iter.Seq[T] {
	return func(yield func(T) bool) {
		next, stop := iter.Pull(b)
		defer stop()
		bv, bok := next()
		for av := range a {
			for bok && compare(bv, av) <= 0 {
				if !yield(bv) {
					return
				}
				bv, bok = next()
			}
			if !yield(av) {
				return
			}
		}
		for bok {
			if !yield(bv) {
				return
			}
			bv, bok = next()
		}
	}
}

// Flatten returns a sequence that yields the elements of each inner sequence in order. It is the inverse of [Chunk].
// The provided sequence is iterated over lazily when the returned sequence is iterated over.
func Flatten[T any](seq iter.Seq[iter.Seq[T]]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for inner := range seq {
			for t := range inner {
				if !yield(t) {
					return
				}
			}
		}
	}
}

// FlattenKV returns a sequence that yields the key-value pairs of each inner sequence in order. It is the inverse of
// [ChunkKV]. The provided sequence is iterated over lazily when the returned sequence is iterated over.
func FlattenKV[K, V any](seq iter.Seq[iter.Seq2[K, V]]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for inner := range seq {
			for k, v := range inner {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// FlatMap maps each value in the sequence to a sequence with the function and yields the elements of each resulting
// sequence in order. Function application happens lazily when the returned sequence is iterated over.
func FlatMap[T, O any](seq iter.Seq[T], fn func(T) iter.Seq[O]) iter.Seq[O] {
	return func(yield func(O) bool) {
		for t := range seq {
			for o := range fn(t) {
				if !yield(o) {
					return
				}
			}
		}
	}
}

// Unique returns a sequence that yields the first occurrence of each distinct value in the sequence. Unlike [Compact],
// which only removes adjacent duplicates, Unique removes duplicates anywhere in the sequence; it needs memory
// proportional to the number of distinct values to do so. The provided sequence is iterated over lazily when the
// returned sequence is iterated over.
func Unique[T comparable](seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for t := range seq {
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			if !yield(t) {
				return
			}
		}
	}
}

// UniqueKV returns a sequence that yields the first occurrence of each distinct key-value pair in the sequence. Unlike
// [CompactKV], which only removes adjacent duplicates, UniqueKV removes duplicates anywhere in the sequence; it needs
// memory proportional to the number of distinct pairs to do so. The provided sequence is iterated over lazily when the
// returned sequence is iterated over.
func UniqueKV[K, V comparable](seq iter.Seq2[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		seen := make(map[KV[K, V]]struct{})
		for k, v := range seq {
			kv := KV[K, V]{K: k, V: v}
			if _, ok := seen[kv]; ok {
				continue
			}
			seen[kv] = struct{}{}
			if !yield(k, v) {
				return
			}
		}
	}
}

// Partition returns two sequences: the first yields the elements for which the function returns true, the second
// yields the rest. Each returned sequence iterates over the provided sequence independently, so iterating both
// iterates the provided sequence twice.
func Partition[T any](seq iter.Seq[T], fn func(T) bool) (iter.Seq[T], iter.Seq[T]) {
	return Filter(seq, fn), DropBy(seq, fn)
}

// PartitionKV returns two sequences: the first yields the key-value pairs for which the function returns true, the
// second yields the rest. Each returned sequence iterates over the provided sequence independently, so iterating both
// iterates the provided sequence twice.
func PartitionKV[K, V any](seq iter.Seq2[K, V], fn func(K, V) bool) (iter.Seq2[K, V], iter.Seq2[K, V]) {
	return FilterKV(seq, fn), DropKVBy(seq, fn)
}

// GroupBy returns a key-value sequence where the keys are the results of applying keyFn to each value and the values
// are slices of the values that produced each key, in encounter order. Keys are yielded in first-seen order. The
// provided sequence is iterated over completely when the returned sequence is iterated over.
func GroupBy[K comparable, T any](seq iter.Seq[T], keyFn func(T) K) iter.Seq2[K, []T] {
	return func(yield func(K, []T) bool) {
		groups := make(map[K][]T)
		var order []K
		for t := range seq {
			k := keyFn(t)
			if _, ok := groups[k]; !ok {
				order = append(order, k)
			}
			groups[k] = append(groups[k], t)
		}
		for _, k := range order {
			if !yield(k, groups[k]) {
				return
			}
		}
	}
}

// Windows returns a sequence of overlapping windows of size consecutive elements. Each window after the first drops
// the oldest element of the previous window and appends the next element of the sequence. If the sequence has fewer
// than size elements the returned sequence is empty. The size must be at least 1; if not, the function will panic. The
// provided sequence is iterated over lazily when the returned sequence is iterated over.
func Windows[T any](seq iter.Seq[T], size int) iter.Seq[iter.Seq[T]] {
	if size < 1 {
		panic("seq: Windows size must be at least 1")
	}
	return func(yield func(iter.Seq[T]) bool) {
		window := make([]T, 0, size)
		for t := range seq {
			if len(window) == size {
				copy(window, window[1:])
				window[size-1] = t
			} else {
				window = append(window, t)
			}
			if len(window) == size {
				w := make([]T, size)
				copy(w, window)
				if !yield(With(w...)) {
					return
				}
			}
		}
	}
}

// WindowsKV is like [Windows] but for key-value pairs. If the sequence has fewer than size pairs the returned sequence
// is empty. The size must be at least 1; if not, the function will panic. The provided sequence is iterated over lazily
// when the returned sequence is iterated over.
func WindowsKV[K, V any](seq iter.Seq2[K, V], size int) iter.Seq[iter.Seq2[K, V]] {
	if size < 1 {
		panic("seq: WindowsKV size must be at least 1")
	}
	return func(yield func(iter.Seq2[K, V]) bool) {
		window := make([]KV[K, V], 0, size)
		for k, v := range seq {
			if len(window) == size {
				copy(window, window[1:])
				window[size-1] = KV[K, V]{K: k, V: v}
			} else {
				window = append(window, KV[K, V]{K: k, V: v})
			}
			if len(window) == size {
				w := make([]KV[K, V], size)
				copy(w, window)
				if !yield(WithKV(w...)) {
					return
				}
			}
		}
	}
}

// All returns true if the function returns true for every value in the sequence. All returns true for an empty
// sequence. The sequence is iterated over until the function returns false when All is called.
func All[T any](seq iter.Seq[T], fn func(T) bool) bool {
	for t := range seq {
		if !fn(t) {
			return false
		}
	}
	return true
}

// AllKV returns true if the function returns true for every key-value pair in the sequence. AllKV returns true for an
// empty sequence. The sequence is iterated over until the function returns false when AllKV is called.
func AllKV[K, V any](seq iter.Seq2[K, V], fn func(K, V) bool) bool {
	for k, v := range seq {
		if !fn(k, v) {
			return false
		}
	}
	return true
}

// None returns true if the function returns false for every value in the sequence. None returns true for an empty
// sequence. This is the opposite of [ContainsFunc]. The sequence is iterated over until the function returns true when
// None is called.
func None[T any](seq iter.Seq[T], fn func(T) bool) bool {
	return !ContainsFunc(seq, fn)
}

// NoneKV returns true if the function returns false for every key-value pair in the sequence. NoneKV returns true for
// an empty sequence. This is the opposite of [ContainsKVFunc]. The sequence is iterated over until the function returns
// true when NoneKV is called.
func NoneKV[K, V any](seq iter.Seq2[K, V], fn func(K, V) bool) bool {
	return !ContainsKVFunc(seq, fn)
}

// Number is the constraint used by the numeric aggregation functions [Sum], [Product], and [Average]. It permits any
// integer or floating point type.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Sum returns the sum of the values in the sequence, or zero if the sequence is empty. The sequence is iterated over
// before Sum returns.
func Sum[T Number](seq iter.Seq[T]) T {
	var sum T
	for t := range seq {
		sum += t
	}
	return sum
}

// Product returns the product of the values in the sequence, or one if the sequence is empty. The sequence is iterated
// over before Product returns.
func Product[T Number](seq iter.Seq[T]) T {
	product := T(1)
	for t := range seq {
		product *= t
	}
	return product
}

// Average returns the arithmetic mean of the values in the sequence. If the sequence is empty, the second return value
// is false. The sequence is iterated over before Average returns.
func Average[T Number](seq iter.Seq[T]) (float64, bool) {
	var sum float64
	var count int
	for t := range seq {
		sum += float64(t)
		count++
	}
	if count == 0 {
		return 0, false
	}
	return sum / float64(count), true
}

// Last returns the final value in the sequence. If the sequence is empty, the second return value is false. The
// sequence is iterated over completely before Last returns.
func Last[T any](seq iter.Seq[T]) (T, bool) {
	var last T
	var found bool
	for t := range seq {
		last = t
		found = true
	}
	return last, found
}

// LastKV returns the final key-value pair in the sequence. If the sequence is empty, the third return value is false.
// The sequence is iterated over completely before LastKV returns.
func LastKV[K, V any](seq iter.Seq2[K, V]) (K, V, bool) {
	var lk K
	var lv V
	var found bool
	for k, v := range seq {
		lk = k
		lv = v
		found = true
	}
	return lk, lv, found
}

// Scan is like [Reduce] but returns a sequence that yields the accumulated value after each element instead of only
// the final value. The initial value itself is not yielded, so the returned sequence has as many elements as the
// provided one. The provided sequence is iterated over lazily when the returned sequence is iterated over.
func Scan[T, O any](seq iter.Seq[T], initial O, fn func(agg O, t T) O) iter.Seq[O] {
	return func(yield func(O) bool) {
		agg := initial
		for t := range seq {
			agg = fn(agg, t)
			if !yield(agg) {
				return
			}
		}
	}
}

// ScanKV is like [ReduceKV] but returns a sequence that yields the accumulated value after each key-value pair instead
// of only the final value. The initial value itself is not yielded, so the returned sequence has as many elements as
// the provided one has pairs. The provided sequence is iterated over lazily when the returned sequence is iterated
// over.
func ScanKV[K, V, O any](seq iter.Seq2[K, V], initial O, fn func(agg O, k K, v V) O) iter.Seq[O] {
	return func(yield func(O) bool) {
		agg := initial
		for k, v := range seq {
			agg = fn(agg, k, v)
			if !yield(agg) {
				return
			}
		}
	}
}

// Cycle returns a sequence that yields the elements of the sequence repeatedly, restarting from the beginning each
// time the provided sequence is exhausted. The returned sequence is infinite unless the provided sequence is empty, so
// bound iteration with something like [Take] or a break. The provided sequence must be re-iterable; single-use
// sequences (like those from [FromChan]) will not restart.
func Cycle[T any](seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			empty := true
			for t := range seq {
				empty = false
				if !yield(t) {
					return
				}
			}
			if empty {
				return
			}
		}
	}
}

// CycleKV returns a sequence that yields the key-value pairs of the sequence repeatedly, restarting from the beginning
// each time the provided sequence is exhausted. The returned sequence is infinite unless the provided sequence is
// empty, so bound iteration with something like [TakeKV] or a break. The provided sequence must be re-iterable;
// single-use sequences will not restart.
func CycleKV[K, V any](seq iter.Seq2[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for {
			empty := true
			for k, v := range seq {
				empty = false
				if !yield(k, v) {
					return
				}
			}
			if empty {
				return
			}
		}
	}
}

// SwapKV returns a sequence with the keys and values of each pair swapped: the values become the keys and the keys
// become the values. The provided sequence is iterated over lazily when the returned sequence is iterated over.
func SwapKV[K, V any](seq iter.Seq2[K, V]) iter.Seq2[V, K] {
	return func(yield func(V, K) bool) {
		for k, v := range seq {
			if !yield(v, k) {
				return
			}
		}
	}
}

// Tap returns a sequence that yields the same elements as the provided sequence, calling the function on each element
// as it passes through. Useful for debugging or other side effects in the middle of a pipeline. The function is
// applied lazily when the returned sequence is iterated over.
func Tap[T any](seq iter.Seq[T], fn func(T)) iter.Seq[T] {
	return func(yield func(T) bool) {
		for t := range seq {
			fn(t)
			if !yield(t) {
				return
			}
		}
	}
}

// TapKV returns a sequence that yields the same key-value pairs as the provided sequence, calling the function on each
// pair as it passes through. Useful for debugging or other side effects in the middle of a pipeline. The function is
// applied lazily when the returned sequence is iterated over.
func TapKV[K, V any](seq iter.Seq2[K, V], fn func(K, V)) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range seq {
			fn(k, v)
			if !yield(k, v) {
				return
			}
		}
	}
}

// FromChanCtx is like [FromChan] but stops when the context is canceled, even if the channel is blocked. The sequence
// ends when the channel is closed or the context is canceled, whichever comes first. Cancellation takes priority: once
// the context is canceled no further values are yielded, even if the channel has values ready.
func FromChanCtx[T any](ctx context.Context, ch <-chan T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			// An already-canceled context must win over a ready channel; a bare select chooses randomly when
			// both cases are ready.
			select {
			case <-ctx.Done():
				return
			default:
			}
			select {
			case <-ctx.Done():
				return
			case t, ok := <-ch:
				if !ok {
					return
				}
				if !yield(t) {
					return
				}
			}
		}
	}
}

// Enumerate returns a key-value sequence that pairs each value in the sequence with its 0-based index. Unlike
// combining [IterKV] with [IntK], the index restarts at 0 each time the returned sequence is iterated over. The
// provided sequence is iterated over lazily when the returned sequence is iterated over.
func Enumerate[T any](seq iter.Seq[T]) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		var i int
		for t := range seq {
			if !yield(i, t) {
				return
			}
			i++
		}
	}
}
