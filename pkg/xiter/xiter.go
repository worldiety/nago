// Package iter provides an adaption of the go iter proposal, see also
// https://go.dev/wiki/RangefuncExperiment. If this proposal is accepted as-is and
// https://github.com/golang/go/issues/46477 is solved, our Seq and Seq2 types will
// be changed to aliases. Signatures will change to always be compatible with
// https://github.com/golang/go/issues/61898 results. E.g. currently order of arguments
// look awkward.
package xiter

import (
	"fmt"
	"iter"
)

func Empty[T any]() iter.Seq[T] {
	return func(yield func(T) bool) {

	}
}

func Empty2[T, V any]() iter.Seq2[T, V] {
	return func(yield func(T, V) bool) {

	}
}

// WithError yields the given error or does nothing if err is nil.
func WithError[T any](err error) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		if err != nil {
			var zero T
			if !yield(zero, err) {
				return
			}
		}
	}
}

func Find[V comparable](it iter.Seq[V], predicate func(V) bool) (V, bool) {
	contains := false
	var res V
	it(func(v V) bool {
		if predicate(v) {
			contains = true
			res = v
			return false
		}

		return true
	})

	return res, contains
}

// Filter returns an iterator over seq that only includes
// the values v for which f(v) is true.
func Filter[V any](f func(V) bool, seq iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		seq(func(v V) bool {
			if f(v) && !yield(v) {
				return false
			}

			return true
		})
	}
}

// Filter2 returns an iterator over seq that only includes
// the pairs k, v for which f(k, v) is true.
func Filter2[K, V any](f func(K, V) bool, seq iter.Seq2[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		seq(func(k K, v V) bool {
			if f(k, v) && !yield(k, v) {
				return false
			}

			return true
		})
	}
}

// Map returns an iterator over f applied to seq.
func Map[In, Out any](f func(In) Out, seq iter.Seq[In]) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		seq(func(in In) bool {
			if !yield(f(in)) {
				return false
			}

			return true
		})
	}
}

// Map2 returns an iterator over f applied to seq.
func Map2[KIn, VIn, KOut, VOut any](f func(KIn, VIn) (KOut, VOut), seq iter.Seq2[KIn, VIn]) iter.Seq2[KOut, VOut] {
	return func(yield func(KOut, VOut) bool) {
		seq(func(in KIn, in2 VIn) bool {
			if !yield(f(in, in2)) {
				return false
			}

			return true
		})
	}
}

// Reduce combines the values in seq using f.
// For each value v in seq, it updates sum = f(sum, v)
// and then returns the final sum.
// For example, if iterating over seq yields v1, v2, v3,
// Reduce returns f(f(f(sum, v1), v2), v3).
func Reduce[Sum, V any](sum Sum, f func(Sum, V) Sum, seq iter.Seq[V]) Sum {
	seq(func(v V) bool {
		sum = f(sum, v)
		return true
	})

	return sum
}

// Reduce2 combines the values in seq using f.
// For each pair k, v in seq, it updates sum = f(sum, k, v)
// and then returns the final sum.
// For example, if iterating over seq yields (k1, v1), (k2, v2), (k3, v3)
// Reduce returns f(f(f(sum, k1, v1), k2, v2), k3, v3).
func Reduce2[Sum, K, V any](sum Sum, f func(Sum, K, V) Sum, seq iter.Seq2[K, V]) Sum {
	seq(func(k K, v V) bool {
		sum = f(sum, k, v)
		return true
	})
	return sum
}

// Limit returns an iterator over seq that stops after n values.
func Limit[V any](seq iter.Seq[V], n int) iter.Seq[V] {
	return func(yield func(V) bool) {
		seq(func(v V) bool {
			if n <= 0 {
				return false
			}

			if !yield(v) {
				return false
			}
			if n--; n <= 0 {
				return true
			}

			return true
		})

	}
}

// Limit2 returns an iterator over seq that stops after n key-value pairs.
func Limit2[K, V any](seq iter.Seq2[K, V], n int) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		seq(func(k K, v V) bool {
			if n <= 0 {
				return false
			}

			if !yield(k, v) {
				return false
			}

			if n--; n <= 0 {
				return true
			}

			return true
		})

	}
}

// BreakOnError stops iteration and sets err on the first err in s
func BreakOnError[K any](err *error, s iter.Seq2[K, error]) iter.Seq[K] {
	return func(yield func(K) bool) {
		s(func(k K, e2 error) bool {
			if e2 != nil {
				*err = e2
				return false
			}

			return yield(k)
		})
	}
}

// Chunks splits and collects the given iter into sub slices. Note that []T slice is owned by the iterator to keep
// allocation rate constant (only 1 allocation of size * sizeof(T)).
// To escape the returned slice, you have to use [slices.Clone].
// Size must not be lower or equal than 0.
func Chunks[T any](seq iter.Seq[T], size int) iter.Seq[[]T] {
	if size < 1 {
		panic(fmt.Errorf("invalid chunk size: %d", size))
	}

	return func(yield func([]T) bool) {
		tmp := make([]T, 0, size)
		for t := range seq {
			tmp = append(tmp, t)
			if len(tmp) == size {
				if !yield(tmp) {
					return
				}
				tmp = tmp[:0]
			}
		}

		// yield whatever fraction has been collected.
		if len(tmp) > 0 {
			yield(tmp)
		}
	}
}

func Zero2[K, V any](values iter.Seq[K]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		var zero V
		for k := range values {
			if !yield(k, zero) {
				return
			}
		}
	}
}
