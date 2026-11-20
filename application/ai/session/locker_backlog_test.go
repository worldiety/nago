// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
)

// lockEntries counts the entries of nago's lock per session of the use cases, read from the maps inside it. It
// observes the field without a signature of its own; if the field is gone, the test says so.
func lockEntries(t *testing.T, uc UseCases) int {
	t.Helper()
	v := reflect.ValueOf(&uc).Elem().FieldByName("locks")
	if !v.IsValid() {
		t.Fatalf("instrument: UseCases has no field locks any more; tell the test where the lock per session lives")
	}
	n, found := mapLens(v, 0)
	if !found {
		t.Fatalf("instrument: the lock per session holds no map; tell the test how to count its entries")
	}
	return n
}

func mapLens(v reflect.Value, depth int) (int, bool) {
	if depth > 6 {
		return 0, false
	}
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return 0, false
		}
		return mapLens(v.Elem(), depth+1)
	case reflect.Map:
		return v.Len(), true
	case reflect.Struct:
		sum, found := 0, false
		for i := range v.NumField() {
			if n, ok := mapLens(v.Field(i), depth+1); ok {
				sum, found = sum+n, true
			}
		}
		return sum, found
	default:
		return 0, false
	}
}

// lockBacklog renames the n sessions s0000 and on, each once, as the system subject; afterwards the lock per
// session is empty. Counter row: during the rename of s0000, held by hold right after its conversation has been
// written, it has one entry, so that the count can see an entry at all.
func lockBacklog(t *testing.T, uc UseCases, n int, hold func(id ID) (reached <-chan struct{}, release func())) {
	t.Helper()
	reached, release := hold("s0000")
	defer release()
	done := make(chan error, 1)
	go func() { done <- uc.Rename(user.SU(), "s0000", "held") }()
	select {
	case <-reached:
	case <-time.After(10 * time.Second):
		t.Fatalf("premise: the rename of s0000 did not write its conversation")
	}
	if got := lockEntries(t, uc); got != 1 {
		t.Errorf("counter row: during a rename the lock per session has %d entries, expected 1", got)
	}
	release()
	if err := <-done; err != nil {
		t.Fatal(err)
	}

	for i := range n {
		if err := uc.Rename(user.SU(), ID(fmt.Sprintf("s%04d", i)), "renamed"); err != nil {
			t.Fatal(err)
		}
	}
	if got := lockEntries(t, uc); got != 0 {
		t.Errorf("after %d renames of different sessions the lock per session has %d entries, expected none", n, got)
	}
}

// holdingRepo holds the save of one session right after it has been written.
type holdingRepo struct {
	Repository
	mu   sync.Mutex
	hold func(id ID)
}

func (r *holdingRepo) Save(s Session) error {
	if err := r.Repository.Save(s); err != nil {
		return err
	}
	r.mu.Lock()
	fn := r.hold
	r.mu.Unlock()
	if fn != nil {
		fn(s.ID)
	}
	return nil
}

// TestLocker_LeavesNothingBehind (wokoda R-DEC-ASSISTANT-SESSION-HEADS, row „1 000 Umbenennungen verschiedener
// Gespräche“, „Entfällt: der Rückstand in nagos Sperre je Gespräch“): after 1 000 renames of different sessions,
// nago's lock per session is empty; counter row: during a rename it has one entry.
func TestLocker_LeavesNothingBehind(t *testing.T) {
	repo := &holdingRepo{Repository: Repository(json.NewSloppyJSONRepository[Session, ID](mem.NewBlobStore(string(Namespace))))}
	for i := range 1000 {
		if err := repo.Repository.Save(Session{ID: ID(fmt.Sprintf("s%04d", i)), Model: "fake-model", CreatedBy: "alice"}); err != nil {
			t.Fatal(err)
		}
	}
	uc := NewUseCases(repo, newTestRDB(t))
	lockBacklog(t, uc, 1000, func(id ID) (<-chan struct{}, func()) {
		reached, release := make(chan struct{}), make(chan struct{})
		var once, freed sync.Once
		repo.mu.Lock()
		defer repo.mu.Unlock()
		repo.hold = func(got ID) {
			if got == id {
				once.Do(func() {
					close(reached)
					<-release
				})
			}
		}
		return reached, func() { freed.Do(func() { close(release) }) }
	})
}
