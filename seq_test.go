package seq

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
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

func ExampleIterIntV() {
	i := With(1, 2, 3, 4)

	s := IterIntV(i)
	for i, v := range s {
		fmt.Printf("%d: %d\n", i, v)
	}

	// Output:
	// 0: 1
	// 1: 2
	// 2: 3
	// 3: 4
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

func ExampleContains() {
	i := With(1, 2, 3, 4, 5)

	fmt.Println(Contains(i, 3))
	fmt.Println(Contains(i, 6))

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

func ExampleReplace() {
	i := With(1, 2, 3, 4, 5)

	i = Replace(i, 2, 6)
	i = Replace(i, 4, 7)

	fmt.Println(slices.Collect(i))

	// Output:
	// [1 6 3 7 5]
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
