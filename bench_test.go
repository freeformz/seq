package seq

import (
	"iter"
	"slices"
	"testing"
)

const benchN = 1000

func benchInts() []int {
	s := make([]int, benchN)
	for i := range s {
		s[i] = i
	}
	return s
}

func benchSeq() iter.Seq[int] {
	return slices.Values(benchInts())
}

func benchSeqKV() iter.Seq2[int, int] {
	return slices.All(benchInts())
}

var (
	sinkInt  int
	sinkBool bool
	sinkKV   KV[int, int]
)

func BenchmarkCompare(b *testing.B) {
	x, y := benchSeq(), benchSeq()
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = Compare(x, y)
	}
}

func BenchmarkCompareKV(b *testing.B) {
	x, y := benchSeqKV(), benchSeqKV()
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = CompareKV(x, y)
	}
}

func BenchmarkEqual(b *testing.B) {
	x, y := benchSeq(), benchSeq()
	b.ReportAllocs()
	for b.Loop() {
		sinkBool = Equal(x, y)
	}
}

func BenchmarkEqualKV(b *testing.B) {
	x, y := benchSeqKV(), benchSeqKV()
	b.ReportAllocs()
	for b.Loop() {
		sinkBool = EqualKV(x, y)
	}
}

func BenchmarkMax(b *testing.B) {
	s := benchSeq()
	b.ReportAllocs()
	for b.Loop() {
		sinkInt, sinkBool = Max(s)
	}
}

func BenchmarkMaxFunc(b *testing.B) {
	s := benchSeq()
	cmp := func(a, b int) int { return a - b }
	b.ReportAllocs()
	for b.Loop() {
		sinkInt, sinkBool = MaxFunc(s, cmp)
	}
}

func BenchmarkMaxFuncKV(b *testing.B) {
	s := benchSeqKV()
	cmp := func(a, b KV[int, int]) int { return a.V - b.V }
	b.ReportAllocs()
	for b.Loop() {
		sinkKV, sinkBool = MaxFuncKV(s, cmp)
	}
}

func BenchmarkMin(b *testing.B) {
	s := benchSeq()
	b.ReportAllocs()
	for b.Loop() {
		sinkInt, sinkBool = Min(s)
	}
}

func BenchmarkMinFunc(b *testing.B) {
	s := benchSeq()
	cmp := func(a, b int) int { return a - b }
	b.ReportAllocs()
	for b.Loop() {
		sinkInt, sinkBool = MinFunc(s, cmp)
	}
}

func BenchmarkCompact(b *testing.B) {
	s := benchSeq()
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = Count(Compact(s))
	}
}

func BenchmarkCompactFunc(b *testing.B) {
	s := benchSeq()
	eq := func(a, b int) bool { return a == b }
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = Count(CompactFunc(s, eq))
	}
}

func BenchmarkIsSorted(b *testing.B) {
	s := benchSeq()
	b.ReportAllocs()
	for b.Loop() {
		sinkBool = IsSorted(s)
	}
}

func BenchmarkDrop(b *testing.B) {
	s := benchSeq()
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = Count(Drop(s, benchN/2))
	}
}

func BenchmarkChunk(b *testing.B) {
	s := benchSeq()
	b.ReportAllocs()
	for b.Loop() {
		for c := range Chunk(s, 16) {
			sinkInt = Count(c)
		}
	}
}

func BenchmarkChunkKV(b *testing.B) {
	s := benchSeqKV()
	b.ReportAllocs()
	for b.Loop() {
		for c := range ChunkKV(s, 16) {
			sinkInt = CountKV(c)
		}
	}
}

func BenchmarkFilter(b *testing.B) {
	s := benchSeq()
	even := func(v int) bool { return v%2 == 0 }
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = Count(Filter(s, even))
	}
}

func BenchmarkMap(b *testing.B) {
	s := benchSeq()
	double := func(v int) int { return v * 2 }
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = Count(Map(s, double))
	}
}
