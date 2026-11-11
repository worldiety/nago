// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"context"
	"sync"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/std/concurrent"
)

type invalidSubject struct{ user.Subject }

func (invalidSubject) Valid() bool { return false }

// TestRotateInvalidatesAuthenticatedPlaintext: a plaintext, which authenticated before the rotation and is served
// from the lookup of the subjects, is invalid right after the rotation, the new one is valid. Before, the old plaintext
// stayed valid in memory until the restart. The same holds for a deletion and for an impersonating token, whose
// subject is the user and not the token.
func TestRotateInvalidatesAuthenticatedPlaintext(t *testing.T) {
	var mutex sync.Mutex
	tokenRepo := json.NewSloppyJSONRepository[Token](mem.NewBlobStore("tokens"))
	repo := data.NewNotifyRepository(nil, Repository(tokenRepo))
	rdb := option.Must(rebac.NewDB(mem.NewBlobStore("rebac")))
	reverse := &concurrent.RWMap[Hash, ID]{}
	subjects := &concurrent.RWMap[Plaintext, authenticated]{}
	repo.AddDeletedObserver(func(_ data.Repository[Token, ID], deleted data.Deleted[ID]) error {
		subjects.DeleteFunc(func(_ Plaintext, a authenticated) bool { return a.token == deleted.ID })
		return nil
	})

	anon := func() user.Subject { return invalidSubject{} }
	impersonated := user.SU()
	fromUser := func(subject permission.Auditable, id user.ID) (option.Opt[user.Subject], error) {
		return option.Some(impersonated), nil
	}

	create := NewCreate(&mutex, repo, user.Argon2IdMin, reverse, rdb)
	authenticate := NewAuthenticateSubject(context.Background(), &mutex, repo, user.Argon2IdMin, reverse, fromUser, subjects, anon, nil, rdb)
	rotate := NewRotate(&mutex, repo, user.Argon2IdMin, reverse, subjects)
	del := NewDelete(&mutex, repo)

	valid := func(p Plaintext) bool {
		t.Helper()
		s, err := authenticate(p)
		if err != nil {
			t.Fatal(err)
		}
		return s.Valid()
	}

	id, old, err := create(user.SU(), CreationData{Name: "rotated"})
	if err != nil {
		t.Fatal(err)
	}

	if !valid(old) || !valid(old) || subjects.Len() != 1 {
		t.Fatalf("premise: the plaintext authenticates and is served from the lookup (%d entries)", subjects.Len())
	}

	fresh, err := rotate(user.SU(), id)
	if err != nil {
		t.Fatal(err)
	}

	if valid(old) {
		t.Errorf("the old plaintext is valid after the rotation")
	}

	if !valid(fresh) {
		t.Errorf("the new plaintext is not valid after the rotation")
	}

	if err := del(user.SU(), id); err != nil {
		t.Fatal(err)
	}

	if valid(fresh) {
		t.Errorf("the plaintext is valid after the deletion")
	}

	// an impersonating token, whose subject is the user
	tok := Token{ID: "imp", Name: "impersonating", Algorithm: user.Argon2IdMin, Impersonation: option.Some(user.ID("someone"))}
	plain := Plaintext("0123456789abcdefghij")
	h, err := plain.TokenHash(user.Argon2IdMin)
	if err != nil {
		t.Fatal(err)
	}
	tok.TokenHash = h
	if err := repo.Save(tok); err != nil {
		t.Fatal(err)
	}
	reverse.Put(HashString(h), tok.ID)

	if !valid(plain) {
		t.Fatalf("premise: the impersonating plaintext authenticates")
	}

	if _, err := rotate(user.SU(), tok.ID); err != nil {
		t.Fatal(err)
	}

	if valid(plain) {
		t.Errorf("the old plaintext of an impersonating token is valid after the rotation")
	}
}
