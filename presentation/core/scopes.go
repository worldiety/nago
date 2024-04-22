package core

import (
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
	"sync"
	"time"
)

// Scopes manages all available scopes and their lifetimes.
type Scopes struct {
	mutex  sync.Mutex
	ticker *time.Ticker
	done   chan bool
	scopes map[ora.ScopeID]*Scope
}

func NewScopes() *Scopes {
	s := &Scopes{ticker: time.NewTicker(time.Minute), done: make(chan bool), scopes: make(map[ora.ScopeID]*Scope)}
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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	scope, ok := s.scopes[id]
	return scope, ok
}

func (s *Scopes) Put(scope *Scope) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.scopes[scope.id] = scope
}

// tick checks all scopes and destroys all scopes which reached EOL.
func (s *Scopes) tick(now time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, scope := range s.scopes {
		if now.After(scope.EOL()) {
			slog.Info("scope is end of life and now destroyed", slog.String("id", string(scope.id)))
			delete(s.scopes, scope.id)
			scope.Destroy()
		}
	}
}

// Destroy stops the internal timer and frees all contained scopes.
func (s *Scopes) Destroy() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.ticker.Stop()

	for _, scope := range s.scopes {
		scope.Destroy()
	}

	clear(s.scopes)
}
