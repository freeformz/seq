package seq

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

func ExampleWith() {
	i := With(1, 2, 3)

	for v := range i {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 2
	// 3
}

func ExampleWithKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	for k, v := range i {
		fmt.Println(k, v)
	}

	// Unordered output:
	// a 1
	// b 2
	// c 3
}

func ExampleMap() {
	i := With(1, 2, 3)

	s := Map(i, strconv.Itoa)
	for v := range s {
		fmt.Printf("%T: %s\n", v, v)
	}

	fmt.Println(slices.Collect(s))

	// Output:
	// string: 1
	// string: 2
	// string: 3
	// [1 2 3]
}

func ExampleMapKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	s := MapKV(i, func(k string, v int) (string, string) {
		return k, "=> " + strconv.Itoa(v)
	})
	for k, v := range s {
		fmt.Println(k, v)
	}

	// Output:
	// a => 1
	// b => 2
	// c => 3
}

func ExampleAppend() {
	i := With(1, 2, 3)

	i = Append(i, 4, 5, 6)
	i = Append(i, 7, 8, 9)
	i = Append(i, 9, 8, 7)

	fmt.Println(slices.Collect(i))

	// Output:
	// [1 2 3 4 5 6 7 8 9 9 8 7]
}

func ExampleAppendKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	i = AppendKV(i, tKV{K: "d", V: 4}, tKV{K: "e", V: 5}, tKV{K: "f", V: 6})
	i = AppendKV(i, tKV{K: "g", V: 7}, tKV{K: "h", V: 8}, tKV{K: "i", V: 9})

	for k, v := range i {
		fmt.Printf("%s%d", k, v)
	}
	fmt.Println()

	// Output:
	// a1b2c3d4e5f6g7h8i9
}

func ExampleFilter() {
	i := With(1, 2, 3, 4, 5, 6, 7, 8, 9)

	s := Filter(i, func(v int) bool {
		return v%2 == 0
	})

	fmt.Println(slices.Collect(s))

	// Output:
	// [2 4 6 8]
}

func ExampleFilterKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	s := FilterKV(i, func(k string, v int) bool {
		return v%2 == 0
	})

	for k, v := range s {
		fmt.Println(k, v)
	}

	// Output:
	// b 2
}

func ExampleMin() {
	i := With(9, 8, 7, 6, 5, 4, 3, 2, 1)

	fmt.Println(Min(i))

	var empty []int
	fmt.Println(Min(slices.Values(empty)))

	// Output:
	// 1 true
	// 0 false
}

func ExampleMinFunc() {
	i := With("hi", "hello", "world")

	fmt.Println(MinFunc(i, strings.Compare))

	var empty []string
	fmt.Println(MinFunc(slices.Values(empty), strings.Compare))

	// Output:
	// hello true
	//  false
}

func ExampleMinFuncKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	fmt.Println(MinFuncKV(i, func(a tKV, b tKV) int {
		return a.V - b.V
	}))

	fmt.Println(MinFuncKV(i, func(a tKV, b tKV) int {
		return strings.Compare(a.K, b.K)
	}))

	fmt.Println(MinFuncKV(i, func(a tKV, b tKV) int {
		if a.V == 3 { // pretend any value of 3 is the min
			return -1
		}
		return 1
	}))

	// Output:
	// {a 1} true
	// {a 1} true
	// {c 3} true
}

func ExampleMax() {
	i := With(9, 8, 7, 6, 5, 4, 3, 2, 1)

	fmt.Println(Max(i))

	var empty []int
	fmt.Println(Max(slices.Values(empty)))

	// Output:
	// 9 true
	// 0 false
}

func ExampleMaxFunc() {
	i := With("hi", "hello", "world")

	fmt.Println(MaxFunc(i, strings.Compare))

	var empty []string
	fmt.Println(MaxFunc(slices.Values(empty), strings.Compare))

	// Output:
	// world true
	//  false
}

func ExampleMaxFuncKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	fmt.Println(MaxFuncKV(i, func(a tKV, b tKV) int {
		return a.V - b.V
	}))

	fmt.Println(MaxFuncKV(i, func(a tKV, b tKV) int {
		return strings.Compare(a.K, b.K)
	}))

	fmt.Println(MaxFuncKV(i, func(a tKV, b tKV) int {
		if a.V == 1 { // pretend any value of 1 is the max
			return 1
		}
		return -1
	}))

	// Output:
	// {c 3} true
	// {c 3} true
	// {a 1} true
}

func ExampleReduce() {
	i := With(1, 2, 3, 4, 5)

	fmt.Println(
		Reduce(i, 10, func(a, b int) int {
			return a + b
		}),
	)

	out := Reduce(i, "a", func(a string, b int) string {
		return strings.Repeat(a, b)
	})
	fmt.Println(out)
	fmt.Println(len(out))

	// Output:
	// 25
	// aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
	// 120
}

func ExampleReduceKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})
	out := ReduceKV(i, "hello: ", func(a, k string, v int) string {
		return a + k + strconv.Itoa(v)
	})
	fmt.Println(out)

	// Output:
	// hello: a1b2c3
}

func ExampleIterKV() {
	i := With(1, 2, 3, 4)

	for i, v := range IterKV(i, IntK[int]()) {
		fmt.Printf("%d: %d\n", i, v)
	}

	for i, v := range IterKV(i, strconv.Itoa) {
		fmt.Printf("%s: %d\n", i, v)
	}

	// Output:
	// 0: 1
	// 1: 2
	// 2: 3
	// 3: 4
	// 1: 1
	// 2: 2
	// 3: 3
	// 4: 4
}

func ExampleIterK() {
	type tKV = KV[string, string]
	i := WithKV(tKV{K: "a", V: "1"}, tKV{K: "b", V: "2"}, tKV{K: "c", V: "3"})
	for k := range IterK(i) {
		fmt.Println(k)
	}

	// Unordered output:
	// a
	// b
	// c
}

func ExampleIterV() {
	type tKV = KV[string, string]
	i := WithKV(tKV{K: "a", V: "1"}, tKV{K: "b", V: "2"}, tKV{K: "c", V: "3"})
	for k := range IterV(i) {
		fmt.Println(k)
	}

	// Unordered output:
	// 1
	// 2
	// 3
}

func ExampleCompact() {
	i := With(1, 2, 2, 3, 3, 4, 5)

	for v := range Compact(i) {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
}

func ExampleCompactFunc() {
	i := With(1, 2, 2, 3, 3, 4, 5)

	for v := range CompactFunc(i, func(a, b int) bool {
		return a == b
	}) {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
}

func ExampleCompactKV() {
	type tKV = KV[string, int]
	i := WithKV(
		tKV{K: "a", V: 1},
		tKV{K: "a", V: 2},
		tKV{K: "a", V: 2},
		tKV{K: "b", V: 3},
		tKV{K: "b", V: 3},
		tKV{K: "c", V: 4},
	)

	for k, v := range CompactKV(i) {
		fmt.Println(k, v)
	}

	// Output:
	// a 1
	// a 2
	// b 3
	// c 4
}

func ExampleCompactKVFunc() {
	type tKV = KV[string, int]
	i := WithKV(
		tKV{K: "a", V: 1},
		tKV{K: "a", V: 2},
		tKV{K: "a", V: 2},
		tKV{K: "b", V: 3},
		tKV{K: "b", V: 3},
		tKV{K: "c", V: 4},
	)

	for k, v := range CompactKVFunc(i, func(a, b tKV) bool {
		return a.K == b.K
	}) {
		fmt.Println(k, v)
	}

	// Output:
	// a 1
	// b 3
	// c 4
}

func ExampleChunk() {
	i := With(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)

	for s := range Chunk(i, 3) {
		fmt.Println(slices.Collect(s))
	}

	// Output:
	// [1 2 3]
	// [4 5 6]
	// [7 8 9]
	// [10 11]
}

func ExampleChunkKV() {
	type tKV = KV[string, int]
	itr := WithKV(
		tKV{K: "a", V: 1},
		tKV{K: "a", V: 2},
		tKV{K: "a", V: 2},
		tKV{K: "b", V: 3},
		tKV{K: "b", V: 3},
		tKV{K: "c", V: 4},
		tKV{K: "c", V: 5},
	)

	var i int
	for chunk := range ChunkKV(itr, 3) {
		fmt.Printf("Chunk %d: ", i)
		for k, v := range chunk {
			fmt.Printf("(%s %d)", k, v)
		}
		fmt.Println()
		i++
	}

	// Output:
	// Chunk 0: (a 1)(a 2)(a 2)
	// Chunk 1: (b 3)(b 3)(c 4)
	// Chunk 2: (c 5)
}

func ExampleCompare() {
	a := With(1, 2, 3)
	b := With(1, 2, 3)
	fmt.Println(Compare(a, b))

	c := With(1, 2)
	fmt.Println(Compare(a, c))

	d := With(1, 2, 4)
	fmt.Println(Compare(a, d))

	e := With(1, 4)
	fmt.Println(Compare(a, e))

	f := With(1, 2, 3, 4)
	fmt.Println(Compare(a, f))

	// Output:
	// 0
	// 1
	// -1
	// -1
	// -1
}

func ExampleCompareFunc() {
	a := With("hi", "hello", "world")
	b := With("hi", "hello", "world")
	fmt.Println(CompareFunc(a, b, strings.Compare))

	c := With("hi", "hello")
	fmt.Println(CompareFunc(a, c, strings.Compare))

	d := With("hi", "hello", "zebra")
	fmt.Println(CompareFunc(a, d, strings.Compare))

	e := With("hi", "zebra")
	fmt.Println(CompareFunc(a, e, strings.Compare))

	f := With("hi", "hello", "world", "zebras")
	fmt.Println(CompareFunc(a, f, strings.Compare))

	// Output:
	// 0
	// 1
	// -1
	// -1
	// -1
}

func ExampleCompareKV() {
	type tKV = KV[string, int]
	a := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})
	b := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})
	fmt.Println(CompareKV(a, b))

	c := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2})
	fmt.Println(CompareKV(a, c))

	d := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 4})
	fmt.Println(CompareKV(a, d))

	e := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "e", V: 3})
	fmt.Println(CompareKV(a, e))

	f := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3}, tKV{K: "d", V: 4})
	fmt.Println(CompareKV(a, f))

	// Output:
	// 0
	// 1
	// -1
	// -1
	// -1
}

func ExampleCompareKVFunc() {
	type aKV = KV[string, int]
	a := WithKV(aKV{K: "a", V: 1}, aKV{K: "b", V: 2}, aKV{K: "c", V: 3})
	b := WithKV(aKV{K: "a", V: 1}, aKV{K: "b", V: 2}, aKV{K: "c", V: 3})
	fmt.Println(CompareKVFunc(a, b, func(a aKV, b aKV) int {
		return a.V - b.V
	}))

	c := WithKV(aKV{K: "a", V: 1}, aKV{K: "b", V: 2})
	fmt.Println(CompareKVFunc(a, c, func(a aKV, b aKV) int {
		return strings.Compare(a.K, b.K)
	}))

	d := WithKV(aKV{K: "a", V: 1}, aKV{K: "b", V: 2}, aKV{K: "c", V: 4})
	fmt.Println(CompareKVFunc(a, d, func(a aKV, b aKV) int {
		return a.V - b.V
	}))

	e := WithKV(aKV{K: "a", V: 1}, aKV{K: "b", V: 2}, aKV{K: "e", V: 3})
	fmt.Println(CompareKVFunc(a, e, func(a aKV, b aKV) int {
		return strings.Compare(a.K, b.K)
	}))

	f := WithKV(aKV{K: "a", V: 1}, aKV{K: "b", V: 2}, aKV{K: "c", V: 3}, aKV{K: "d", V: 4})
	fmt.Println(CompareKVFunc(a, f, func(a aKV, b aKV) int {
		return a.V - b.V
	}))

	type bKV = KV[string, string]
	g := WithKV(bKV{K: "a", V: "1"}, bKV{K: "b", V: "2"}, bKV{K: "c", V: "3"})
	fmt.Println(CompareKVFunc(a, g, func(a aKV, b bKV) int {
		if c := strings.Compare(a.K, b.K); c != 0 {
			return c
		}
		return strings.Compare(strconv.Itoa(a.V), b.V)
	}))

	// Output:
	// 0
	// 1
	// -1
	// -1
	// -1
	// 0
}

func ExampleContains() {
	i := With(1, 2, 3, 4, 5)

	fmt.Println(Contains(i, 3))
	fmt.Println(Contains(i, 6))

	// Output:
	// true
	// false
}

func ExampleContainsKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	fmt.Println(ContainsKV(i, "b", 2))
	fmt.Println(ContainsKV(i, "d", 1))

	// Output:
	// true
	// false
}

func ExampleContainsFunc() {
	i := With("hi", "hello", "world")

	fmt.Println(ContainsFunc(i, func(s string) bool { return s == "hello" }))
	fmt.Println(ContainsFunc(i, func(s string) bool { return s == "zebra" }))

	// Output:
	// true
	// false
}

func ExampleContainsKVFunc() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	fmt.Println(ContainsKVFunc(i, func(k string, v int) bool { return k == "b" && v == 2 }))
	fmt.Println(ContainsKVFunc(i, func(k string, v int) bool { return k == "d" && v == 1 }))

	// Output:
	// true
	// false
}

func ExampleEqual() {
	a := With(1, 2, 3)
	b := With(1, 2, 3)
	fmt.Println(Equal(a, b))

	c := With(1, 2)
	fmt.Println(Equal(a, c))

	d := With(3, 2, 1)
	fmt.Println(Equal(a, d))

	// Output:
	// true
	// false
	// false
}

func ExampleEqualKV() {
	type tKV = KV[string, int]
	a := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})
	b := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})
	fmt.Println(EqualKV(a, b))

	c := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2})
	fmt.Println(EqualKV(a, c))

	d := WithKV(tKV{K: "c", V: 3}, tKV{K: "b", V: 2}, tKV{K: "a", V: 1})
	fmt.Println(EqualKV(a, d))

	e := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3}, tKV{K: "d", V: 4})
	fmt.Println(EqualKV(a, e))

	// Output:
	// true
	// false
	// false
	// false
}

func ExampleEqualFunc() {
	a := With("hi", "hello", "world")
	b := With("hi", "hello", "world")
	fmt.Println(EqualFunc(a, b, strings.EqualFold))

	c := With("hi", "hello")
	fmt.Println(EqualFunc(a, c, strings.EqualFold))

	d := With("hi", "hello", "zebra")
	fmt.Println(EqualFunc(a, d, strings.EqualFold))

	e := With("hi", "hello", "WORLD")
	fmt.Println(EqualFunc(a, e, strings.EqualFold))

	// Output:
	// true
	// false
	// false
	// true
}

func ExampleEqualKVFunc() {
	type tKV = KV[string, int]
	a := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})
	b := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})
	fmt.Println(EqualKVFunc(a, b, func(a tKV, b tKV) bool {
		return a.V == b.V && a.K == b.K
	}))

	c := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2})
	fmt.Println(EqualKVFunc(a, c, func(a tKV, b tKV) bool {
		return a.V == b.V && a.K == b.K
	}))

	d := WithKV(tKV{K: "c", V: 3}, tKV{K: "b", V: 2}, tKV{K: "a", V: 1})
	fmt.Println(EqualKVFunc(a, d, func(a tKV, b tKV) bool {
		return a.V == b.V && a.K == b.K
	}))

	e := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3}, tKV{K: "d", V: 4})
	fmt.Println(EqualKVFunc(a, e, func(a tKV, b tKV) bool {
		return a.V == b.V && a.K == b.K
	}))

	f := WithKV(tKV{K: "A", V: 1}, tKV{K: "B", V: 2}, tKV{K: "C", V: 3})
	fmt.Println(EqualKVFunc(a, f, func(a tKV, b tKV) bool {
		return a.V == b.V && strings.EqualFold(a.K, b.K)
	}))

	// Output:
	// true
	// false
	// false
	// false
	// true
}

func ExampleRepeat() {
	i := Repeat(3, "hi")
	for v := range i {
		fmt.Println(v)
	}

	// Output:
	// hi
	// hi
	// hi
}

func ExampleRepeatKV() {
	i := RepeatKV(3, "a", 1)
	for k, v := range i {
		fmt.Println(k, v)
	}

	// Output:
	// a 1
	// a 1
	// a 1
}

func ExampleReplace() {
	i := With(1, 2, 3, 4, 5)

	i = Replace(i, 2, 6)
	i = Replace(i, 4, 7)

	fmt.Println(slices.Collect(i))

	// Output:
	// [1 6 3 7 5]
}

func ExampleReplaceKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	i = ReplaceKV(i, tKV{"a", 1}, tKV{"a", 6})
	i = ReplaceKV(i, tKV{"c", 7}, tKV{"c", 8}) // no effect

	for k, v := range i {
		fmt.Println(k, v)
	}
	fmt.Println()

	// Output:
	// a 6
	// b 2
	// c 3
}

func ExampleIsSorted() {
	i := With(1, 2, 3, 4, 5)
	fmt.Println(IsSorted(i))
	fmt.Println(IsSorted(i))

	j := With(1, 2, 3, 4, 3)
	fmt.Println(IsSorted(j))

	// Output:
	// true
	// true
	// false
}

func ExampleIsSortedKV() {
	type kv = KV[string, int]
	i := WithKV(kv{K: "a", V: 1}, kv{K: "b", V: 2}, kv{K: "c", V: 3})
	fmt.Println(IsSortedKV(i))

	i = WithKV(kv{K: "a", V: 1}, kv{K: "b", V: 2}, kv{K: "b", V: 2}, kv{K: "c", V: 3})
	fmt.Println(IsSortedKV(i))

	i = WithKV(kv{K: "a", V: 1}, kv{K: "b", V: 2}, kv{K: "c", V: 3}, kv{K: "d", V: 2})
	fmt.Println(IsSortedKV(i))

	i = WithKV(kv{"b", 1}, kv{"a", 2}, kv{"c", 3})
	fmt.Println(IsSortedKV(i))

	// Output:
	// true
	// true
	// false
	// false
}

func ExampleFromChan() {
	values := make(chan int, 3)
	go func() {
		for i := range 10 {
			values <- i
		}
		close(values)
	}()

	vals := slices.Collect(FromChan(values))
	fmt.Println(vals)

	// Output:
	// [0 1 2 3 4 5 6 7 8 9]
}

func ExampleToChan() {
	i := With(1, 2, 3, 4, 5)
	ch := ToChan(i)

	for v := range ch {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
}

func ExampleToChanCtx() {
	i := With(1, 2, 3, 4, 5)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := ToChanCtx(ctx, i)

	for v := range ch {
		fmt.Println(v)
		if v == 3 {
			cancel()
		}
	}

	// Output:
	// 1
	// 2
	// 3
}

func ExampleCoalesce() {
	i := With(0, 0, 4, 5)

	fmt.Println(Coalesce(i))

	// Output:
	// 4 true
}

func ExampleCoalesceKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 0}, tKV{K: "b", V: 0}, tKV{K: "c", V: 4}, tKV{K: "d", V: 5})

	fmt.Println(CoalesceKV(i))

	// Output:
	// {c 4} true
}

func ExampleCount() {
	i := With(1, 2, 3, 4)

	fmt.Println(Count(i))

	// Output:
	// 4
}

func ExampleCountKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	fmt.Println(CountKV(i))

	// Output:
	// 3
}

func ExampleCountBy() {
	i := With(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	fmt.Println(CountBy(i, func(v int) bool {
		return v%2 == 0
	}))

	// Output:
	// 5
}

func ExampleCountKVBy() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	fmt.Println(CountKVBy(i, func(k string, v int) bool {
		return v%2 == 0
	}))

	// Output:
	// 1
}

func ExampleCountValues() {
	i := With(1, 1, 2, 2, 3, 3, 3, 4)

	for k, v := range CountValues(i) {
		fmt.Printf("%d: %v\n", k, v)
	}

	// Unordered output:
	// 1: 2
	// 2: 2
	// 3: 3
	// 4: 1
}

func ExampleDrop() {
	i := With(1, 2, 3, 4, 5)

	for v := range Drop(i, 2) {
		fmt.Println(v)
	}

	for v := range Drop(i, 0) {
		fmt.Println(v)
	}

	// doesn't print anything
	for v := range Drop(i, 100) {
		fmt.Println(v)
	}

	// Output:
	// 3
	// 4
	// 5
	// 1
	// 2
	// 3
	// 4
	// 5
}

func ExampleDropKV() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	for k, v := range DropKV(i, 2) {
		fmt.Println(k, v)
	}

	for k, v := range DropKV(i, 0) {
		fmt.Println(k, v)
	}

	// doesn't print anything
	for k, v := range DropKV(i, 100) {
		fmt.Println(k, v)
	}

	// Output:
	// c 3
	// a 1
	// b 2
	// c 3
}

func ExampleDropBy() {
	i := With(1, 2, 3, 4, 5, 6, 7, 8, 9)

	s := DropBy(i, func(v int) bool {
		return v%2 == 0
	})

	fmt.Println(slices.Collect(s))

	// Output:
	// [1 3 5 7 9]
}

func ExampleDropKVBy() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	s := DropKVBy(i, func(k string, v int) bool {
		return v%2 == 0
	})

	for k, v := range s {
		fmt.Println(k, v)
	}

	// Output:
	// a 1
	// c 3
}

func ExampleEveryUntil() {
	var i int
	for t := range EveryUntil(time.Millisecond, time.Now().Add(10*time.Millisecond)) {
		_ = t // 2025-03-23 18:53:05.064589166 -0700 PDT m=+0.007687209
		i++
	}

	fmt.Println(i)

	// Output:
	// 9
}

func ExampleEveryN() {
	var i int
	for t := range EveryN(time.Millisecond, 10) {
		_ = t // 2025-03-23 18:53:05.064589166 -0700 PDT m=+0.007687209
		i++
	}

	fmt.Println(i)

	// Output:
	// 10
}

func ExampleMapToKV() {
	i := With(1, 2, 3)

	for k, v := range MapToKV(i, func(i int) (string, int) {
		return string([]byte{byte(64 + i)}), i
	}) {
		fmt.Println(k, v)
	}

	// Output:
	// A 1
	// B 2
	// C 3
}

func ExampleFind() {
	i := With(1, 2, 3, 4, 5)

	fmt.Println(Find(i, 3))

	fmt.Println(Find(i, 6))

	// Output:
	// 2 true
	// 5 false
}

func ExampleFindBy() {
	i := With(1, 2, 3, 4, 5)

	v, idx, ok := FindBy(i, func(v int) bool {
		return v == 3
	})

	fmt.Println(v, idx, ok)

	v, idx, ok = FindBy(i, func(v int) bool {
		return v == 6
	})

	fmt.Println(v, idx, ok)

	// Output:
	// 3 2 true
	// 0 5 false
}

func ExampleFindByKey() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	fmt.Println(FindByKey(i, "b"))

	fmt.Println(FindByKey(i, "d"))

	// Output:
	// 2 1 true
	// 0 3 false
}

func ExampleFindByValue() {
	type tKV = KV[string, int]
	i := WithKV(tKV{K: "a", V: 1}, tKV{K: "b", V: 2}, tKV{K: "c", V: 3})

	fmt.Println(FindByValue(i, 2))

	fmt.Println(FindByValue(i, 4))

	// Output:
	// b 1 true
	//  3 false
}
