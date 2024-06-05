package core

import "log/slog"

type ViewRoot interface {
	// AddDestroyObserver registers an observer which is called, before the root component of the window is destroyed.
	AddDestroyObserver(fn func()) (removeObserver func())

	// Invalidate renders the tree and sends it to the actual frontend for displaying. Usually you should not use
	// this directly, because the request-response cycles triggers this automatically. However, if backend
	// data has changed due to other domain events, you have to notify the view tree to redraw and potentially
	// to load the data again from repositories. In those cases you likely want to use [core.Iterable.Iter] to
	// always rebuild the entire tree from the according property.
	Invalidate()
}

type scopeViewRoot struct {
	observers map[int]func()
	hnd       int
	scope     *Scope
	component Component
	destroyed bool
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
	s.destroyed = true
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

	if s.destroyed {
		return
	}
	evt := s.scope.render(0, s.component)
	s.scope.Publish(evt)
}

func (s *scopeViewRoot) setComponent(component Component) {
	s.component = component
}
