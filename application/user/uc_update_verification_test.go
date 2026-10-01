// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"sync"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/permission"
	datamem "go.wdy.de/nago/pkg/data/mem"
)

// The verified flag is applied as given, instead of always verifying.
func TestUpdateVerificationHonorsFlag(t *testing.T) {
	repo := &datamem.Repository[User, ID]{}
	if err := repo.Save(User{ID: "u1", Email: "a@example.com", VerificationCode: NewCode(time.Hour)}); err != nil {
		t.Fatal(err)
	}

	var mutex sync.Mutex
	admin := testSubject{id: "admin", perms: []permission.ID{PermUpdateOtherContact}}
	update := NewUpdateVerification(&mutex, repo)
	updateByMail := NewUpdateVerificationByMail(&mutex, repo, func(subject permission.Auditable, mail Email) (option.Opt[User], error) {
		return repo.FindByID("u1")
	})

	if err := update(admin, "u1", true); err != nil {
		t.Fatal(err)
	}

	usr := option.Must(repo.FindByID("u1")).Unwrap()
	if !usr.EMailVerified || usr.VerificationCode.Value != "" {
		t.Fatalf("expected a verified user without code, got %+v", usr)
	}

	if err := updateByMail(admin, "a@example.com", false); err != nil {
		t.Fatal(err)
	}

	usr = option.Must(repo.FindByID("u1")).Unwrap()
	if usr.EMailVerified || usr.VerificationCode.Value == "" || !usr.RequiresVerification() {
		t.Fatalf("expected an unverified user with a fresh code, got %+v", usr)
	}
}
