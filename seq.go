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

// Map the values in the sequence to a new sequence of values by applying the function fn to each value. Function application
// happens lazily when the returned sequence is iterated over.
func Map[T, O any](seq iter.Seq[T], fn func(T) O) iter.Seq[O] {
	return func(yield func(o O) bool) {
		seq(func(t T) bool {
			return yield(fn(t))
		})
	}
}

// Append the items to the sequence and return an extended sequence. The provided sequence and appended items are iterated over
// lazily when the returned sequence is iterated over.
func Append[T any](seq iter.Seq[T], items ...T) iter.Seq[T] {
	return func(yield func(T) bool) {
		seq(yield)
		for _, item := range items {
			if !yield(item) {
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

// Iter2 converts an iter.Seq[T] to an iter.Seq2[int, T]. The provided sequence is iterated over lazily when the returned
// sequence is iterated over.
func Iter2[T any](iter iter.Seq[T]) func(func(i int, t T) bool) {
	var i int
	return func(yield func(i int, t T) bool) {
		for t := range iter {
			if !yield(i, t) {
				return
			}
			i++
		}
	}
}

// Max value of the sequence. Uses max builtin to compare values. The second value is false if the sequence is empty. The
// sequence is iterated over before Max returns.
func Max[T cmp.Ordered](seq iter.Seq[T]) (T, bool) {
	var mt T
	var value bool
	for i, t := range Iter2(seq) {
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
	for i, t := range Iter2(seq) {
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
	for i, t := range Iter2(seq) {
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
	for i, t := range Iter2(seq) {
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

// Compact returns an iterator that yields all values that are not equal to the previous value. The provided sequence is iterated
// over lazily when the returned sequence is iterated over.
func Compact[T comparable](seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		var prev T
		for i, t := range Iter2(seq) {
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
		for i, t := range Iter2(seq) {
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
	for i, t := range Iter2(seq) {
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
