// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"errors"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
)

// An application deletes what it keeps per conversation together with the session, and a failure keeps the
// session, so that nothing is orphaned.
func TestOnSessionDelete(t *testing.T) {
	var mgmt Management
	nagotest.New(t, func(c *application.Configurator) {
		c.SetApplicationID("de.worldiety.ai.sessiondeletetest")
		mgmt = std.Must(Enable(c))
	})

	var deleted []session.ID
	var fail error
	mgmt.OnSessionDelete(func(id session.ID) error {
		if fail != nil {
			return fail
		}
		deleted = append(deleted, id)
		return nil
	})

	su := user.SU()
	s, err := mgmt.SessionUseCases.Create(su, session.CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	fail = errors.New("cannot delete the workbook")
	if err := mgmt.SessionUseCases.Delete(su, s.ID); err == nil {
		t.Fatal("expected the deletion to fail")
	}

	if opt, _ := mgmt.SessionUseCases.FindByID(su, s.ID); opt.IsNone() {
		t.Fatal("the session must be kept, if the application cannot delete its data")
	}

	fail = nil
	if err := mgmt.SessionUseCases.Delete(su, s.ID); err != nil {
		t.Fatal(err)
	}

	if len(deleted) != 1 || deleted[0] != s.ID {
		t.Fatalf("expected the hook for the session, got %v", deleted)
	}
}
