package core

type ViewRoot interface {
	// AddDestroyObserver registers an observer which is called, before the root component of the window is destroyed.
	AddDestroyObserver(fn func()) (removeObserver func())

	// Invalidate renders the current view root and sends it to the server.
	Invalidate()
}

type scopeViewRoot struct {
	observers map[int]func()
	hnd       int
}

func newScopeViewRoot() *scopeViewRoot {
	return &scopeViewRoot{observers: map[int]func(){}}
}

func (s *scopeViewRoot) AddDestroyObserver(fn func()) (removeObserver func()) {
	s.hnd++
	myHnd := s.hnd
	s.observers[myHnd] = fn
	return func() {
		delete(s.observers, myHnd)
	}
}

func (s *scopeViewRoot) Destroy() {
	for _, f := range s.observers {
		f()
	}
	clear(s.observers)
}

func (s *scopeViewRoot) Invalidate() {
	// TODO implement me
}
