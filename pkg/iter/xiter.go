// Package iter provides an adaption of the go iter proposal, see also
// https://go.dev/wiki/RangefuncExperiment. If this proposal is accepted as-is and
// https://github.com/golang/go/issues/46477 is solved, our Seq and Seq2 types will
// be changed to aliases.
package iter

type Seq[V any] func(yield func(V) bool)
type Seq2[K, V any] func(yield func(K, V) bool)

// Filter returns an iterator over seq that only includes
// the values v for which f(v) is true.
func Filter[V any](f func(V) bool, seq Seq[V]) Seq[V] {
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
func Filter2[K, V any](f func(K, V) bool, seq Seq2[K, V]) Seq2[K, V] {
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
func Map[In, Out any](f func(In) Out, seq Seq[In]) Seq[Out] {
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
func Map2[KIn, VIn, KOut, VOut any](f func(KIn, VIn) (KOut, VOut), seq Seq2[KIn, VIn]) Seq2[KOut, VOut] {
	return func(yield func(KOut, VOut) bool) {
		seq(func(in KIn, in2 VIn) bool {
			if !yield(f(in, in2)) {
				return false
			}

			return true
		})
	}
}
