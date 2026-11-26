// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"sync"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/presentation/proto"
)

// Scopes manages all available scopes and their lifetimes.
type Scopes struct {
	eolTicker    *time.Ticker
	eolDone      chan bool
	updateTicker *time.Ticker
	updateDone   chan bool
	scopes       concurrent.CoWMap[proto.ScopeID, *Scope]
	destroyed    atomic.Bool
	// mutex serializes the structural changes, so that a scope which is connected is never reaped at the same time.
	mutex sync.Mutex
}

func NewScopes(fps int) *Scopes {
	s := &Scopes{
		eolTicker:    time.NewTicker(time.Minute),
		eolDone:      make(chan bool),
		updateTicker: time.NewTicker(time.Duration(1000/fps) * time.Millisecond),
		updateDone:   make(chan bool),
	}
	go func() {
		for {
			select {
			case <-s.eolDone:
				return
			case t := <-s.eolTicker.C:
				s.tick(t)
			}
		}

	}()

	go func() {
		for {
			select {
			case <-s.updateDone:
				return
			case t := <-s.updateTicker.C:
				s.updateTick(t)
			}
		}
	}()

	return s
}

func (s *Scopes) Get(id proto.ScopeID) (*Scope, bool) {
	scope, ok := s.scopes.Get(id)
	return scope, ok
}

// getOrCreate returns the living scope of the given id and keeps it alive. Otherwise, a new scope is created
// and put.
func (s *Scopes) getOrCreate(id proto.ScopeID, create func() *Scope) *Scope {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if scope, ok := s.scopes.Get(id); ok && !scope.destroyed.Load() {
		scope.Tick()
		return scope
	}

	scope := create()
	s.scopes.Put(scope.id, scope)
	return scope
}

// connectOwned returns the living scope of the given id, if the given session owns it, and keeps it alive. An
// unknown id gets a new scope, which the session owns from then on. A scope of another session is left untouched
// and reported as false, so that a refused connection neither keeps it alive nor changes anything about it.
func (s *Scopes) connectOwned(id proto.ScopeID, owner session.ID, create func() *Scope) (*Scope, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if scope, ok := s.scopes.Get(id); ok && !scope.destroyed.Load() {
		if !scope.ownedBy(owner) {
			return nil, false
		}

		scope.Tick()
		return scope, true
	}

	scope := create()
	scope.owner.Store(&owner)
	s.scopes.Put(scope.id, scope)
	return scope, true
}

// Remove removes and destroys the scope of the given id. It returns false, if no such scope exists.
func (s *Scopes) Remove(id proto.ScopeID) bool {
	s.mutex.Lock()
	scope, ok := s.scopes.Get(id)
	if ok {
		s.scopes.Delete(id)
	}
	s.mutex.Unlock()

	if ok {
		scope.Destroy()
	}

	return ok
}

// tick checks all scopes and destroys all scopes which reached EOL.
func (s *Scopes) tick(now time.Time) {
	s.scopes.Each(func(key proto.ScopeID, scope *Scope) bool {
		if !now.After(scope.EOL()) {
			return true
		}

		// check again under lock, because the scope may have been connected in the meantime
		s.mutex.Lock()
		current, ok := s.scopes.Get(key)
		expired := ok && current == scope && now.After(scope.EOL())
		if expired {
			s.scopes.Delete(key)
		}
		s.mutex.Unlock()

		if expired {
			scope.Destroy()
		}

		return true
	})
}

func (s *Scopes) updateTick(now time.Time) {
	s.scopes.Each(func(key proto.ScopeID, scope *Scope) bool {
		scope.updateTick(now)
		return true
	})

}

// Destroy stops the internal timer and frees all contained scopes.
func (s *Scopes) Destroy() {
	if !s.destroyed.CompareAndSwap(false, true) {
		return
	}

	s.eolTicker.Stop()
	s.updateTicker.Stop()
	close(s.eolDone)
	close(s.updateDone)

	s.scopes.Each(func(key proto.ScopeID, scope *Scope) bool {
		scope.Destroy()
		return true
	})

	s.scopes.Clear()
}
