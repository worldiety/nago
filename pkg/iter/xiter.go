// Package iter provides an adaption of the go iter proposal, see also
// https://go.dev/wiki/RangefuncExperiment. If this proposal is accepted as-is and
// https://github.com/golang/go/issues/46477 is solved, our Seq and Seq2 types will
// be changed to aliases.
package iter

type Seq[V any] func(yield func(V) bool)
type Seq2[K, V any] func(yield func(K, V) bool)

// Values returns an iterator over the values in the slice,
// starting with s[0].
func Values[Slice ~[]Elem, Elem any](s Slice) Seq[Elem] {
	return func(yield func(Elem) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
		return
	}
}
