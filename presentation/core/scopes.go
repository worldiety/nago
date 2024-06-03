package core

import (
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
	"time"
)

// Scopes manages all available scopes and their lifetimes.
type Scopes struct {
	ticker    *time.Ticker
	done      chan bool
	scopes    concurrent.CoWMap[ora.ScopeID, *Scope]
	destroyed concurrent.Value[bool]
}

func NewScopes() *Scopes {
	s := &Scopes{ticker: time.NewTicker(time.Minute), done: make(chan bool)}
	go func() {
		for {
			select {
			case <-s.done:
				return
			case t := <-s.ticker.C:
				s.tick(t)
			}
		}

	}()
	return s
}

func (s *Scopes) Get(id ora.ScopeID) (*Scope, bool) {
	scope, ok := s.scopes.Get(id)
	return scope, ok
}

func (s *Scopes) Put(scope *Scope) {
	s.scopes.Put(scope.id, scope)
}

// tick checks all scopes and destroys all scopes which reached EOL.
func (s *Scopes) tick(now time.Time) {
	s.scopes.Each(func(key ora.ScopeID, scope *Scope) bool {
		if now.After(scope.EOL()) {
			slog.Info("scope is end of life and now destroyed", slog.String("id", string(scope.id)))
			s.scopes.Delete(scope.id)
			scope.Destroy()
		}

		return true
	})

}

// Destroy stops the internal timer and frees all contained scopes.
func (s *Scopes) Destroy() {
	if !concurrent.CompareAndSwap(&s.destroyed, false, true) {
		return
	}

	s.ticker.Stop()

	s.scopes.Each(func(key ora.ScopeID, scope *Scope) bool {
		scope.Destroy()
		return true
	})

	s.scopes.Clear()
}
