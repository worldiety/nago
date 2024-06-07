package core

import (
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
)

type ViewRoot interface {
	Factory() ora.ComponentFactoryId

	// AddDestroyObserver registers an observer which is called, before the root component of the window is destroyed.
	AddDestroyObserver(fn func()) (removeObserver func())

	// AddWindowChangedObserver registers an observer to be called, after the frontend has adjusted its size
	// at least in a significant way. Frontends are free to optimize, e.g. they may send pixel exact events
	// or only when the size class or a media break point was changed.
	AddWindowChangedObserver(fn func()) (removeObserver func())

	// AddWindowSizeClassObserver registers an observer which is always called if the size class changes.
	AddWindowSizeClassObserver(fn func(sizeClass ora.WindowSizeClass)) (removeObserver func())

	// Invalidate renders the tree and sends it to the actual frontend for displaying. Usually you should not use
	// this directly, because the request-response cycles triggers this automatically. However, if backend
	// data has changed due to other domain events, you have to notify the view tree to redraw and potentially
	// to load the data again from repositories. In those cases you likely want to use [core.Iterable.Iter] to
	// always rebuild the entire tree from the according property.
	Invalidate()
}

type scopeViewRoot struct {
	destroyObservers       map[int]func()
	windowChangedObservers map[int]func()
	hnd                    int
	scope                  *Scope
	component              Component
	destroyed              bool
	scopeWindow            *scopeWindow
}

func (s *scopeViewRoot) Factory() ora.ComponentFactoryId {
	return s.scopeWindow.factory
}

func newScopeViewRoot(scope *Scope) *scopeViewRoot {
	return &scopeViewRoot{
		destroyObservers: map[int]func(){}, scope: scope,
		windowChangedObservers: map[int]func(){},
	}
}

func (s *scopeViewRoot) AddWindowSizeClassObserver(fn func(sizeClass ora.WindowSizeClass)) (removeObserver func()) {
	sizeClass := s.scope.windowInfo.SizeClass()

	s.hnd++
	myHnd := s.hnd
	s.windowChangedObservers[myHnd] = func() {
		newClass := s.scope.windowInfo.SizeClass()
		if sizeClass != newClass {
			sizeClass = newClass
			fn(sizeClass)
			s.Invalidate()
		}
	}

	return func() {
		delete(s.windowChangedObservers, myHnd)
	}
}

func (s *scopeViewRoot) AddWindowChangedObserver(fn func()) (removeObserver func()) {
	s.hnd++
	myHnd := s.hnd
	s.windowChangedObservers[myHnd] = func() {
		fn()
		s.Invalidate()
	}
	return func() {
		delete(s.windowChangedObservers, myHnd)
	}
}

func (s *scopeViewRoot) AddDestroyObserver(fn func()) (removeObserver func()) {
	s.hnd++
	myHnd := s.hnd
	s.destroyObservers[myHnd] = fn
	return func() {
		delete(s.destroyObservers, myHnd)
	}
}

func (s *scopeViewRoot) onWindowUpdated() {
	for _, f := range s.windowChangedObservers {
		f()
	}
}

func (s *scopeViewRoot) Destroy() {
	s.destroyed = true
	for _, f := range s.destroyObservers {
		f()
	}
	clear(s.destroyObservers)
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
