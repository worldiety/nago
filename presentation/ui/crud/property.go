package crud

type Property[E any, T any] interface {
	Set(dst *E, v T)
	Get(*E) T
}

func Ptr[E, T any](f func(model *E) *T) Property[E, T] {
	return fieldPtr[E, T](f)
}

type fieldPtr[E any, T any] func(model *E) *T

func (f fieldPtr[E, T]) Set(dst *E, v T) {
	fieldPtr := f(dst)
	*fieldPtr = v
}

func (f fieldPtr[E, T]) Get(e *E) T {
	fieldPtr := f(e)
	return *fieldPtr
}
