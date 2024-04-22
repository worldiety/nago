package core

import "log/slog"

type ViewRoot interface {
	// AddDestroyObserver registers an observer which is called, before the root component of the window is destroyed.
	AddDestroyObserver(fn func()) (removeObserver func())

	// Invalidate renders the current view root and sends it to the server.
	Invalidate()
}

type scopeViewRoot struct {
	observers map[int]func()
	hnd       int
	scope     *Scope
	component Component
}

func newScopeViewRoot(scope *Scope) *scopeViewRoot {
	return &scopeViewRoot{observers: map[int]func(){}, scope: scope}
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
	s.component = nil
}

func (s *scopeViewRoot) Invalidate() {
	if s.component == nil {
		slog.Error("cannot invalidate nil component. This may be either early in construction cycle or after destruction.")
		return
	}

	evt := s.scope.render(0, s.component)
	s.scope.Publish(evt)
}

func (s *scopeViewRoot) setComponent(component Component) {
	s.component = component
}
