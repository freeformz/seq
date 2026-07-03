# Sequence (Iterator) Utilities for Golang

![ci status](https://github.com/freeformz/seq/actions/workflows/ci.yaml/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/freeformz/seq)](https://goreportcard.com/report/github.com/freeformz/seq)
[![GoDoc](https://godoc.org/github.com/freeformz/seq?status.svg)](http://godoc.org/github.com/freeformz/seq)

Golang's "missing" iterator/sequence functions.

## Construction Functions

### iter.Seq[T]

* `With(...T) iter.Seq[T]`: Construct a sequence using the provided values
* `FromChan(<-chan T) iter.Seq[T]`: Returns a sequence that produces values until the channel is closed
* `FromChanCtx(context.Context, <-chan T) iter.Seq[T]`: Like FromChan but also stops when the context is canceled
* `Repeat(int, T) iter.Seq[T]`: Returns a sequence which repeats the value n times

### iter.Seq2[K,V]

* `WithKV(...KV[K,V]) iter.Seq2[K,V]`: Construct a key-value sequence using the provided key-values
* `RepeatKV(int, K, V) iter.Seq2[K,V]`: Returns a sequence which repeats the key-value pair n times

## Conversion Functions

* `ToChan(iter.Seq[T]) <-chan T`: Returns a channel that produces values until the sequence is exhausted
* `ToChanCtx(context.Context, iter.Seq[T]) <-chan T`: Returns a channel that produces values until the sequence is exhausted or the context is canceled
* `IterKV(iter.Seq[V], func(V) K) iter.Seq2[K,V]`: Converts an iter.Seq[V] to an iter.Seq2[K,V] using keyFn for keys
* `IterK(iter.Seq2[K,V]) iter.Seq[K]`: Converts an iter.Seq2[K,V] to an iter.Seq[K] (keys only)
* `IterV(iter.Seq2[K,V]) iter.Seq[V]`: Converts an iter.Seq2[K,V] to an iter.Seq[V] (values only)
* `MapToKV(iter.Seq[T], func(T) (K,V)) iter.Seq2[K,V]`: Maps values to key-value pairs
* `SwapKV(iter.Seq2[K,V]) iter.Seq2[V,K]`: Swaps the keys and values of each pair
* `Enumerate(iter.Seq[T]) iter.Seq2[int,T]`: Pairs each value with its 0-based index; the index restarts on each iteration

## Transformation Functions

### Mapping

* `Map(iter.Seq[T], func(T) O) iter.Seq[O]`: Maps the items in the sequence to another type
* `MapKV(iter.Seq2[K,V], func(K,V) (K1,V1)) iter.Seq2[K1,V1]`: Maps the key-value pairs to other types
* `FlatMap(iter.Seq[T], func(T) iter.Seq[O]) iter.Seq[O]`: Maps each value to a sequence and yields the elements of each in order
* `Scan(iter.Seq[T], O, func(O,T) O) iter.Seq[O]`: Like Reduce but lazily yields the accumulated value after each element
* `ScanKV(iter.Seq2[K,V], O, func(O,K,V) O) iter.Seq[O]`: Like ReduceKV but lazily yields the accumulated value after each pair
* `Tap(iter.Seq[T], func(T)) iter.Seq[T]`: Yields the same elements, calling the function on each as it passes through
* `TapKV(iter.Seq2[K,V], func(K,V)) iter.Seq2[K,V]`: Yields the same pairs, calling the function on each as it passes through

### Filtering

* `Filter(iter.Seq[T], func(T) bool) iter.Seq[T]`: Filter values by applying fn to each value
* `FilterKV(iter.Seq2[K,V], func(K,V) bool) iter.Seq2[K,V]`: Filter key-value pairs by applying fn to each pair

### Appending

* `Append(iter.Seq[T], ...T) iter.Seq[T]`: Returns a new sequence with additional items appended
* `AppendKV(iter.Seq2[K,V], ...KV[K,V]) iter.Seq2[K,V]`: Returns a new sequence with additional key-value pairs appended

### Combining

* `Concat(...iter.Seq[T]) iter.Seq[T]`: Yields the elements of each sequence in order
* `ConcatKV(...iter.Seq2[K,V]) iter.Seq2[K,V]`: Yields the key-value pairs of each sequence in order
* `Zip(iter.Seq[A], iter.Seq[B]) iter.Seq2[A,B]`: Pairs the elements of two sequences positionally, ending at the shorter one
* `Merge(iter.Seq[T], iter.Seq[T]) iter.Seq[T]`: Merges two sorted sequences into one sorted sequence
* `MergeFunc(iter.Seq[T], iter.Seq[T], func(T,T) int) iter.Seq[T]`: Like Merge but uses a comparison function

### Cycling

* `Cycle(iter.Seq[T]) iter.Seq[T]`: Repeats the sequence forever (empty input yields an empty sequence)
* `CycleKV(iter.Seq2[K,V]) iter.Seq2[K,V]`: Repeats the key-value sequence forever (empty input yields an empty sequence)

### Replacement

* `Replace(iter.Seq[T], old, new T) iter.Seq[T]`: Replace old values with new values
* `ReplaceKV(iter.Seq2[K,V], old, new KV[K,V]) iter.Seq2[K,V]`: Replace old key-value pairs with new ones

### Compacting

* `Compact(iter.Seq[T]) iter.Seq[T]`: Yields all values that are not equal to the previous value
* `CompactFunc(iter.Seq[T], func(T,T) bool) iter.Seq[T]`: Like Compact but uses a function to compare elements
* `CompactKV(iter.Seq2[K,V]) iter.Seq2[K,V]`: Yields all key-value pairs that are not equal to the previous pair
* `CompactKVFunc(iter.Seq2[K,V], func(KV[K,V], KV[K,V]) bool) iter.Seq2[K,V]`: Like CompactKV but uses a function to compare pairs
* `Unique(iter.Seq[T]) iter.Seq[T]`: Yields the first occurrence of each distinct value (removes duplicates anywhere, not just adjacent)
* `UniqueKV(iter.Seq2[K,V]) iter.Seq2[K,V]`: Yields the first occurrence of each distinct key-value pair

### Chunking

* `Chunk(iter.Seq[T], int) iter.Seq[iter.Seq[T]]`: Chunk the sequence into chunks of specified size
* `ChunkKV(iter.Seq2[K,V], int) iter.Seq[iter.Seq2[K,V]]`: Chunk key-value pairs into chunks of specified size
* `Windows(iter.Seq[T], int) iter.Seq[iter.Seq[T]]`: Overlapping windows of the specified size (sliding by one element)
* `WindowsKV(iter.Seq2[K,V], int) iter.Seq[iter.Seq2[K,V]]`: Overlapping windows of key-value pairs
* `Flatten(iter.Seq[iter.Seq[T]]) iter.Seq[T]`: Yields the elements of each inner sequence in order (the inverse of Chunk)
* `FlattenKV(iter.Seq[iter.Seq2[K,V]]) iter.Seq2[K,V]`: Yields the key-value pairs of each inner sequence in order (the inverse of ChunkKV)

### Grouping

* `GroupBy(iter.Seq[T], func(T) K) iter.Seq2[K,[]T]`: Groups values by key in first-seen order
* `Partition(iter.Seq[T], func(T) bool) (iter.Seq[T], iter.Seq[T])`: Splits into matching and non-matching sequences
* `PartitionKV(iter.Seq2[K,V], func(K,V) bool) (iter.Seq2[K,V], iter.Seq2[K,V])`: Splits key-value pairs into matching and non-matching sequences

### Taking

* `Take(iter.Seq[T], int) iter.Seq[T]`: Take the first n elements of the sequence
* `TakeKV(iter.Seq2[K,V], int) iter.Seq2[K,V]`: Take the first n key-value pairs of the sequence
* `TakeWhile(iter.Seq[T], func(T) bool) iter.Seq[T]`: Take leading elements while the function returns true
* `TakeKVWhile(iter.Seq2[K,V], func(K,V) bool) iter.Seq2[K,V]`: Take leading key-value pairs while the function returns true

### Dropping

* `Drop(iter.Seq[T], int) iter.Seq[T]`: Drop n elements from the start of the sequence
* `DropKV(iter.Seq2[K,V], int) iter.Seq2[K,V]`: Drop n key-value pairs from the start of the sequence
* `DropBy(iter.Seq[T], func(T) bool) iter.Seq[T]`: Drop all elements for which the function returns true
* `DropKVBy(iter.Seq2[K,V], func(K,V) bool) iter.Seq2[K,V]`: Drop all key-value pairs for which the function returns true
* `DropWhile(iter.Seq[T], func(T) bool) iter.Seq[T]`: Drop leading elements while the function returns true, then yield the rest
* `DropKVWhile(iter.Seq2[K,V], func(K,V) bool) iter.Seq2[K,V]`: Drop leading key-value pairs while the function returns true, then yield the rest

## Aggregation Functions

### Min/Max

* `Min(iter.Seq[T]) (T, bool)`: Min value from the sequence using built-in comparison
* `MinFunc(iter.Seq[T], func(T,T) int) (T, bool)`: Min value using a comparison function
* `MinFuncKV(iter.Seq2[K,V], func(KV[K,V], KV[K,V]) int) (KV[K,V], bool)`: Min key-value pair using a comparison function
* `Max(iter.Seq[T]) (T, bool)`: Max value from the sequence using built-in comparison
* `MaxFunc(iter.Seq[T], func(T,T) int) (T, bool)`: Max value using a comparison function
* `MaxFuncKV(iter.Seq2[K,V], func(KV[K,V], KV[K,V]) int) (KV[K,V], bool)`: Max key-value pair using a comparison function

### Reduction

* `Reduce(iter.Seq[T], O, func(O,T) O) O`: Reduce the sequence to a single value
* `ReduceKV(iter.Seq2[K,V], O, func(O,K,V) O) O`: Reduce key-value pairs to a single value

### Numeric

* `Sum(iter.Seq[T]) T`: Sum of the values (zero for an empty sequence); T is any integer or float type
* `Product(iter.Seq[T]) T`: Product of the values (one for an empty sequence); T is any integer or float type
* `Average(iter.Seq[T]) (float64, bool)`: Arithmetic mean of the values; false if the sequence is empty

### Counting

* `Count(iter.Seq[T]) int`: Returns the number of elements in the sequence
* `CountKV(iter.Seq2[K,V]) int`: Returns the number of key-value pairs in the sequence
* `CountBy(iter.Seq[T], func(T) bool) int`: Count elements for which the function returns true
* `CountKVBy(iter.Seq2[K,V], func(K,V) bool) int`: Count key-value pairs for which the function returns true
* `CountValues(iter.Seq[T]) iter.Seq2[T,int]`: Returns a sequence where keys are values and values are their counts

## Comparison Functions

* `Compare(iter.Seq[T], iter.Seq[T]) int`: Compare two sequences using cmp.Compare
* `CompareFunc(iter.Seq[T], iter.Seq[T], func(T,T) int) int`: Compare two sequences using a comparison function
* `CompareKV(iter.Seq2[K,V], iter.Seq2[K,V]) int`: Compare two key-value sequences using cmp.Compare
* `CompareKVFunc(iter.Seq2[AK,AV], iter.Seq2[BK,BV], func(KV[AK,AV], KV[BK,BV]) int) int`: Compare two key-value sequences using a comparison function

## Equality Functions

* `Equal(iter.Seq[T], iter.Seq[T]) bool`: Returns true if sequences are equal
* `EqualKV(iter.Seq2[K,V], iter.Seq2[K,V]) bool`: Returns true if key-value sequences are equal
* `EqualFunc(iter.Seq[T], iter.Seq[T], func(T,T) bool) bool`: Test equality using a comparison function
* `EqualKVFunc(iter.Seq2[AK,AV], iter.Seq2[BK,BV], func(KV[AK,AV], KV[BK,BV]) bool) bool`: Test key-value equality using a comparison function

## Search Functions

### Contains

* `Contains(iter.Seq[T], T) bool`: Returns true if the value is in the sequence
* `ContainsKV(iter.Seq2[K,V], K, V) bool`: Returns true if the key-value pair is in the sequence
* `ContainsFunc(iter.Seq[T], func(T) bool) bool`: Returns true if predicate returns true for any value
* `ContainsKVFunc(iter.Seq2[K,V], func(K,V) bool) bool`: Returns true if predicate returns true for any key-value pair

### Predicates

* `All(iter.Seq[T], func(T) bool) bool`: Returns true if the function returns true for every value (true for empty)
* `AllKV(iter.Seq2[K,V], func(K,V) bool) bool`: Returns true if the function returns true for every key-value pair (true for empty)
* `None(iter.Seq[T], func(T) bool) bool`: Returns true if the function returns false for every value (true for empty)
* `NoneKV(iter.Seq2[K,V], func(K,V) bool) bool`: Returns true if the function returns false for every key-value pair (true for empty)

### Finding

* `Find(iter.Seq[T], T) (int, bool)`: Returns the index of the first occurrence of the value
* `FindBy(iter.Seq[T], func(T) bool) (T, int, bool)`: Returns the first value for which the function returns true
* `FindByKey(iter.Seq2[K,V], K) (V, int, bool)`: Returns the value of the first key-value pair with the given key
* `FindByValue(iter.Seq2[K,V], V) (K, int, bool)`: Returns the key of the first key-value pair with the given value
* `At(iter.Seq[T], int) (T, bool)`: Returns the value at the given 0-based index, or zero value and false if out of range
* `AtKV(iter.Seq2[K,V], int) (K, V, bool)`: Returns the key and value at the given 0-based index, or zero values and false if out of range
* `Last(iter.Seq[T]) (T, bool)`: Returns the final value in the sequence, or zero value and false if empty
* `LastKV(iter.Seq2[K,V]) (K, V, bool)`: Returns the final key-value pair in the sequence, or zero values and false if empty

## Utility Functions

* `Coalesce(iter.Seq[T]) (T, bool)`: Returns the first non-zero value in the sequence
* `CoalesceKV(iter.Seq2[K,V]) (KV[K,V], bool)`: Returns the first key-value pair with a non-zero value
* `IsSorted(iter.Seq[T]) bool`: Returns true if the sequence is sorted
* `IsSortedKV(iter.Seq2[K,V]) bool`: Returns true if the key-value sequence is sorted
* `IntK() func(V) int`: Returns a function that generates increasing integers starting at 0

## Time-based Functions

* `EveryUntil(time.Duration, time.Time) iter.Seq[time.Time]`: Yields time every duration until the specified time
* `EveryN(time.Duration, int) iter.Seq[time.Time]`: Yields time every duration for n times

## Types

* `KV[K,V]`: A struct that pairs a key and value together for use with key-value sequence functions
* `Number`: A constraint permitting any integer or floating point type, used by Sum, Product, and Average
