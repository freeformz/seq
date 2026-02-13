# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go library (`github.com/freeformz/seq`) providing functional iterator/sequence utilities built on Go's `iter.Seq[T]` and `iter.Seq2[K,V]` types. Requires Go 1.24+. Zero external dependencies. Single package, single source file (`seq.go`).

## Commands

```bash
# Test (with race detector)
go test -v -race ./...

# Lint
go vet ./...
go install honnef.co/go/tools/cmd/staticcheck@latest && staticcheck ./...
```

## Architecture & Conventions

**Dual API pattern**: Most functions have two variants — one for `iter.Seq[T]` and a `KV` suffix version for `iter.Seq2[K,V]`. The `KV[K,V]` struct bridges between the two.

**Comparison function variants**: Functions involving ordering have up to three forms — constrained to `cmp.Ordered`, a `Func` version accepting a comparator, and a `FuncKV` version.

**Lazy vs eager**: Transformation functions (Map, Filter, Chunk, Drop, etc.) return new iterators via closures over `yield func(T) bool`. Aggregation functions (Reduce, Min, Max, Count, etc.) consume the entire sequence eagerly.

**Testing**: All tests are `Example` functions — they serve as both documentation and regression tests. No traditional unit tests. Run a single example with `go test -run ExampleFunctionName`.

**Commit tags**: Commit messages use `#patch`, `#minor`, `#major`, or `#none` suffixes for automated semantic version bumping.
