package iter

// deprecated
// see https://github.com/golang/go/issues/61897#issuecomment-1790799275, due by probably Go 1.23
type Seq[V any] func(yield func(V) bool)
