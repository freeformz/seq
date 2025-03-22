package seq

import (
	"cmp"
	"iter"
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

// IterIntV converts an iter.Seq[T] to an iter.Seq2[int, T]. The provided sequence is iterated over lazily when the returned
// sequence is iterated over.
func IterIntV[T any](iter iter.Seq[T]) iter.Seq2[int, T] {
	i := -1
	return IterKV(iter, func(T) int {
		i++
		return i
	})
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
	for i, t := range IterIntV(seq) {
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
	for i, t := range IterIntV(seq) {
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

// Min value from the sequence. Uses min built in to compare values. The second value is false if the sequence is empty. The
// sequence is iterated over before Min returns.
func Min[T cmp.Ordered](seq iter.Seq[T]) (T, bool) {
	var mt T
	var value bool
	for i, t := range IterIntV(seq) {
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
	for i, t := range IterIntV(seq) {
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
		for i, t := range IterIntV(seq) {
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
		for i, t := range IterIntV(seq) {
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

// Contains returns true if the value is in the sequence. The sequence is iterated over when Contains is called.
func Contains[T comparable](seq iter.Seq[T], value T) bool {
	for t := range seq {
		if t == value {
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

// EqualFunc is like [Equal] but uses the function to compare elements.
func EqualFunc[T any](a, b iter.Seq[T], equal func(T, T) bool) bool {
	return CompareFunc(a, b, func(a, b T) int {
		if equal(a, b) {
			return 0
		}
		return -1
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

// IsSorted returns true if the sequence is sorted. The provided sequence is iterated over before IsSorted returns.
func IsSorted[T cmp.Ordered](seq iter.Seq[T]) bool {
	var prev T
	for i, t := range IterIntV(seq) {
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
