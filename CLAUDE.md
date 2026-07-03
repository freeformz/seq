# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go library (`github.com/freeformz/seq`) providing functional iterator/sequence utilities built on Go's `iter.Seq[T]` and `iter.Seq2[K,V]` types. Requires Go 1.25+. Zero external dependencies. Single package, single source file (`seq.go`).

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

**Testing**: All tests in the main package are `Example` functions — they serve as both documentation and regression tests. No traditional unit tests in the main package. Run a single example with `go test -run ExampleFunctionName`. The `stresstest` subpackage is the exception: it holds regular `Test` functions for behaviors that can't be expressed as Examples (panics, hang regressions, data races, goroutine leaks) and should be run with `-race`.

**Commit tags**: `.github/workflows/release.yaml` runs on every PR merged into `main`. It scans the squashed merge commit for a `#major`, `#minor`, `#patch`, or `#none` token, bumps a `vX.Y.Z` tag accordingly, and publishes a matching GitHub Release. This repo only allows squash merges, and GitHub's squash settings here (`COMMIT_OR_PR_TITLE` / `COMMIT_MESSAGES`) mean the scanned text is the PR title (when the PR has multiple commits) plus the full text of every individual commit in the PR — so a tag placed on any one commit, or on the PR title, is picked up. If several different tokens appear, the highest-ranking one wins (`major` > `minor` > `patch`); `#none` skips the bump entirely regardless of the others. **This repo overrides the action's default bump to `patch`** (not `minor`), so an untagged PR still cuts a real release — always tag deliberately rather than relying on the default.

Pick the tag by the scope of the change, not the size of the diff:
- `#major` — breaking changes to the public API: removing/renaming an exported identifier, changing a signature or documented behavior incompatibly.
- `#minor` — backward-compatible additions to the public API (new exported functions/types), and module-level changes that only take effect via a new published version (e.g. a `go.mod` `retract` directive, bumping the `go` directive).
- `#patch` — bug fixes and internal changes that don't add or break API surface.
- `#none` — changes that shouldn't cut a release at all: docs-only, test-only, or CI/tooling-only changes (e.g. editing `release.yaml` itself) that don't affect the published module.

Use exactly one tag per PR reflecting its most significant change; put it in the PR title so it stays correct however the commits get squashed.

**Landmine**: the scanner does a plain substring match over the full squashed commit text — it has no idea a token appears inside prose rather than as a directive. Never write the literal, hash-prefixed tokens together in a commit message body when merely describing the convention (e.g. "supports #major/#minor/#patch/#none") — an incidental `#major` substring there will out-rank real `#minor`/`#patch` tags elsewhere in the same PR and trigger an unintended major release (this happened once: PR #9's docs commit describing this exact convention accidentally cut `v1.0.0` from a `v0.4.0` base). When discussing the tokens in a commit message, drop the `#` prefix (`major`/`minor`/`patch`/`none`) or reference this file instead. The tokens are safe to write in full inside file content (e.g. this doc) — only commit *messages* are scanned, not diffs.
