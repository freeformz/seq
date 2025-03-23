package seq

import (
	"cmp"
	"iter"
	"sync/atomic"
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
func IterKV[K, V any](iter iter.Seq[V], keyFn func(V) K) iter.Seq2[K, V] {
	return func(yield func(k K, v V) bool) {
		for v := range iter {
			k := keyFn(v)
			if !yield(k, v) {
				return
			}
		}
	}
}

// IterK converts an iter.Seq2[K, V] to an iter.Seq[K]. The provided sequence is iterated over lazily when the returned
// sequence is iterated over.
func IterK[K, V any](iter iter.Seq2[K, V]) iter.Seq[K] {
	return func(yield func(k K) bool) {
		for k := range iter {
			if !yield(k) {
				return
			}
		}
	}
}

// IterV converts an iter.Seq2[K, V] to an iter.Seq[V]. The provided sequence is iterated over lazily when the returned
// sequence is iterated over.
func IterV[K, V any](iter iter.Seq2[K, V]) iter.Seq[V] {
	return func(yield func(v V) bool) {
		for _, v := range iter {
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
	for i, t := range IterKV(seq, IntK[T]()) {
		switch i {
		case 0:
			mt = t
			value = true
		default:
			mt = max(t, mt)
		}
	}
	return mt, value
}

// MaxFunc is like [Max] but uses the function to compare elements. The provided sequence is iterated over before MaxFunc returns.
func MaxFunc[T any](seq iter.Seq[T], compare func(T, T) int) (T, bool) {
	var mt T
	var value bool
	for i, t := range IterKV(seq, IntK[T]()) {
		switch i {
		case 0:
			mt = t
			value = true
		default:
			if compare(t, mt) > 0 {
				mt = t
			}
		}
	}
	return mt, value
}

// MaxFuncKV is like [MaxFunc] but for key-value pairs. The provided sequence is iterated over before MaxFuncKV returns.
func MaxFuncKV[K, V any](seq iter.Seq2[K, V], compare func(KV[K, V], KV[K, V]) int) (KV[K, V], bool) {
	var mt KV[K, V]
	var value bool
	var i int
	for k, v := range seq {
		switch i {
		case 0:
			mt = KV[K, V]{K: k, V: v}
			value = true
		default:
			t := KV[K, V]{K: k, V: v}
			if compare(t, mt) > 0 {
				mt = t
			}
		}
		i++
	}
	return mt, value
}

// Min value from the sequence. Uses min built in to compare values. The second value is false if the sequence is empty. The
// sequence is iterated over before Min returns.
func Min[T cmp.Ordered](seq iter.Seq[T]) (T, bool) {
	var mt T
	var value bool
	for i, t := range IterKV(seq, IntK[T]()) {
		switch i {
		case 0:
			mt = t
			value = true
		default:
			mt = min(t, mt)
		}
	}
	return mt, value
}

// MinFunc is like [Min] but uses the function to compare elements. The provided sequence is iterated over before MinFunc returns.
func MinFunc[T any](seq iter.Seq[T], compare func(T, T) int) (T, bool) {
	var mt T
	var value bool
	for i, t := range IterKV(seq, IntK[T]()) {
		switch i {
		case 0:
			mt = t
			value = true
		default:
			if compare(t, mt) < 0 {
				mt = t
			}
		}
	}
	return mt, value
}

// MinFuncKV is like [MinFunc] but for key-value pairs. The provided sequence is iterated over before MinFuncKV returns.
func MinFuncKV[K, V any](seq iter.Seq2[K, V], compare func(KV[K, V], KV[K, V]) int) (KV[K, V], bool) {
	var mt KV[K, V]
	var value bool
	var i int
	for k, v := range seq {
		switch i {
		case 0:
			mt = KV[K, V]{K: k, V: v}
			value = true
		default:
			t := KV[K, V]{K: k, V: v}
			if compare(t, mt) < 0 {
				mt = t
			}
		}
		i++
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
		for i, t := range IterKV(seq, IntK[T]()) {
			switch i {
			case 0:
				prev = t
				if !yield(t) {
					return
				}
			default:
				if prev != t {
					prev = t
					if !yield(t) {
						return
					}
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
		for i, t := range IterKV(seq, IntK[T]()) {
			switch i {
			case 0:
				prev = t
				if !yield(t) {
					return
				}
			default:
				if !equal(prev, t) {
					prev = t
					if !yield(t) {
						return
					}
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
		for k, v := range seq {
			if prev.K != k || prev.V != v {
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
		for k, v := range seq {
			if !equal(prev, KV[K, V]{K: k, V: v}) {
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
// over. The last chunk may have fewer than size elements.
func Chunk[T any](seq iter.Seq[T], size int) iter.Seq[iter.Seq[T]] {
	return func(yield func(iter.Seq[T]) bool) {
		var chunk []T
		for t := range seq {
			chunk = append(chunk, t)
			if len(chunk) == size {
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
// iterated over. The last chunk may have fewer than size elements.
func ChunkKV[K, V any](seq iter.Seq2[K, V], size int) iter.Seq[iter.Seq2[K, V]] {
	return func(yield func(iter.Seq2[K, V]) bool) {
		var chunk []KV[K, V]
		for k, v := range seq {
			chunk = append(chunk, KV[K, V]{K: k, V: v})
			if len(chunk) == size {
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
// elements is returned. If both sequences are equal until one of them ends, the shorter sequence is considered less
// than the longer one. The result is 0 if a == b, -1 if a < b, and +1 if a > b.
func CompareFunc[T any](a, b iter.Seq[T], compare func(T, T) int) int {
	bvals := make(chan T)
	exit := make(chan struct{})
	defer close(exit)

	go func() {
		defer close(bvals)
		for v := range b {
			select {
			case bvals <- v:
			case <-exit:
				return
			}
		}
	}()

	for av := range a {
		bv, ok := <-bvals
		if !ok { // b is shorter than a
			return 1
		}
		if c := compare(av, bv); c != 0 {
			return c
		}
	}

	// done with a, check if b is longer
	// if bvals isn't closed b is longer than a
	if _, ok := <-bvals; ok {
		return -1
	}

	// a and b are equal
	return 0
}

// CompareKV is like [CompareKVFunc] but uses the cmp.Compare function to compare keys and values.
func CompareKV[K, V cmp.Ordered](a, b iter.Seq2[K, V]) int {
	return CompareKVFunc(a, b, func(a, b KV[K, V]) int {
		if cmp.Compare(a.K, b.K) == 0 {
			return cmp.Compare(a.V, b.V)
		}
		return cmp.Compare(a.K, b.K)
	})
}

// CompareKVFunc compares the key-value pairs of a and b, using the compare func on each pair of key-value pairs. The key-value
// pairs are compared sequentially, until one key-value pair is not equal to the other. The result of comparing the first
// non-matching key-value pairs is returned. If both sequences are equal until one of them ends, the shorter sequence is
// considered less than the longer one. The result is 0 if a == b, -1 if a < b, and +1 if a > b.
func CompareKVFunc[AK, AV, BK, BV any](a iter.Seq2[AK, AV], b iter.Seq2[BK, BV], compare func(a KV[AK, AV], b KV[BK, BV]) int) int {
	bvals := make(chan KV[BK, BV])
	exit := make(chan struct{})
	defer close(exit)

	go func() {
		defer close(bvals)
		for k, v := range b {
			select {
			case bvals <- KV[BK, BV]{k, v}:
			case <-exit:
				return
			}
		}
	}()

	for ak, av := range a {
		bv, ok := <-bvals
		if !ok { // b is shorter than a
			return 1
		}
		if c := compare(KV[AK, AV]{ak, av}, bv); c != 0 {
			return c
		}
	}

	// done with a, check if b is longer
	// if bvals isn't closed b is longer than a
	if _, ok := <-bvals; ok {
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

// ContainsFunc returns true if the function returns true for any value in the sequence. The sequence is iterated over when
// ContainsFunc is called.
func ContainsFunc[T any](seq iter.Seq[T], equal func(T) bool) bool {
	for t := range seq {
		if equal(t) {
			return true
		}
	}
	return false
}

// ContainsKVFunc returns true if the function returns true for any key-value pair in the sequence. The sequence is iterated over
// when ContainsKVFunc is called.
func ContainsKVFunc[K, V any](seq iter.Seq2[K, V], equal func(K, V) bool) bool {
	for k, v := range seq {
		if equal(k, v) {
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
		for i := 0; i < n; i++ {
			if !yield(t) {
				return
			}
		}
	}
}

// RepeatKV returns a sequence which repeats the key-value pair n times.
func RepeatKV[K, V any](n int, k K, v V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for i := 0; i < n; i++ {
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

// IsSorted returns true if the sequence is sorted. The provided sequence is iterated over before IsSorted returns. [cmp.Compare]
// // is used to compare elements.
func IsSorted[T cmp.Ordered](seq iter.Seq[T]) bool {
	var prev T
	for i, t := range IterKV(seq, IntK[T]()) {
		switch i {
		case 0:
			prev = t
		default:
			if cmp.Compare(t, prev) < 0 {
				return false
			}
			prev = t
		}
	}
	return true
}

// IsSortedKV returns true if the sequence is sorted. The provided sequence is iterated over before IsSortedKV returns.
// [cmp.Compare] is used to compare keys and values
func IsSortedKV[K, V cmp.Ordered](seq iter.Seq2[K, V]) bool {
	var prev KV[K, V]
	for k, v := range seq {
		if (cmp.Compare(k, prev.K) < 0) || (cmp.Compare(v, prev.V) < 0) {
			return false
		}
		prev.K = k
		prev.V = v
	}
	return true
}
