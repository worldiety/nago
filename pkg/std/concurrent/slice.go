package concurrent

import "sync"

// CoWSlice is a copy-on-write slice.
type CoWSlice[T any] struct {
	mutex sync.Mutex
	slice []T
}

// Len does not allocate. Note, that Len does not make much sense in concurrent situations.
func (l *CoWSlice[T]) Len() int {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return len(l.slice)
}

// Append locks and copies the entire set. This is very expensive.
func (l *CoWSlice[T]) Append(v ...T) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	tmp := make([]T, len(l.slice), len(l.slice)+len(v))
	copy(tmp, l.slice)
	tmp = append(tmp, v...)
	l.slice = tmp
}

func (l *CoWSlice[T]) Clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.slice = nil
}

// Each iterates over all items. This is cheap and does not allocate and can never deadlock. Append or Clear will
// allocate new slices underneath, so that Each always iterates on an immutable copy.
func (l *CoWSlice[T]) Each(yield func(T) bool) {
	l.mutex.Lock()
	ref := l.slice
	l.mutex.Unlock()

	for _, t := range ref {
		if !yield(t) {
			return
		}
	}
}
