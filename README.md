# Sequence (Iterator) Utilities for Golang

![ci status](https://github.com/freeformz/seq/actions/workflows/ci.yaml/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/freeformz/seq)](https://goreportcard.com/report/github.com/freeformz/seq)
[![GoDoc](https://godoc.org/github.com/freeformz/seq?status.svg)](http://godoc.org/github.com/freeformz/seq)

Golang's "missing" iterator/sequence functions.

## iter.Seq helpers

* `With(...T) iter.Seq[T]` : Construct a sequence using the provided values;
* `FromChan(<-chan) iter.Seq[T]`: Returns a sequence that produces values until the channel is closed;
* `ToChan(iter.Seq[T]) <-chan`: Returns a channel that produces values until the sequence is exhausted;
* `ToChanCtx(context.Context, iter.Seq[T]) <-chan`: Returns a channel that produces values until the sequence is
  exhausted or the context is canceled;
* `Map(iter.Seq[T], func(T) O) iter.Seq[O]`: Maps the items in the sequence to another type;
* `Append(iter.Seq[T], ...T) iter.Seq[T]`: Returns a new sequence that includes the items from the passed sequence, plus
  the additional items;
* `Filter(iter.Seq[T], func(T) bool) iter.Seq[T]`: Filter the values in the sequence by applying fn to each value;
* `IterKV(iter.Seq[T], func(V) K) iter.Seq2[K,V]`: ...

## iter.Seq2 helpers

Some of these helpers use seq.KV, some do not. I've generally tried to use seq.KV when it would:

1. easily be confusing to deal with keys and values as separate values
2. when they have to be paired together to avoid having to handle odd numbers of input

* `WithKV(...KV[K,V])` : Construct a key-value (iter.Seq2) sequence using the provided key-values;
* `MapKV(iter.Seq2[K,V], func(K,V) (K1,V1)) iter.Seq2[K1,V2]`: Maps the items in the sequence to other types;
* `AppendKV(iter.Seq2[K,V], ...KV[K,V]) iter.Seq2[K,V]`: Returns a new sequence that includes the items from the passed sequence, plus
  the additional KV pairs;
* `FilterKV(iter.Seq2[K,V], func(K,V) bool) iter.Seq2[K,V]`: Filter the values in the sequence by applying fn to each value;
