// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfglibrary

import (
	"slices"
	"testing"

	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/permission"
)

// TestLibrarianRolePermissions is the regression that a real production defect calls for: the bootstrap
// account holds every permission by construction, so nothing notices when a shipped role grants none of them
// and the first real user simply sees an empty page.
//
// The identifiers are spelled out as literals rather than read from the code under test. A test that derives
// its expectation from what it checks agrees with every change, including the one that removes a permission.
func TestLibrarianRolePermissions(t *testing.T) {
	want := []permission.ID{
		"tutorial.library.book.find_all",
		"tutorial.library.book.lend",
		"tutorial.library.book.return",
	}

	got := slices.Clone(LibrarianPermissions())
	slices.Sort(got)
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestAssistantPermissionsAreNotInTheDomainRole documents the split: the framework permissions belong to the
// role cfgai ships, so an operator can hand out the assistant independently of what a user may do here.
func TestAssistantPermissionsAreNotInTheDomainRole(t *testing.T) {
	domain := LibrarianPermissions()

	for _, pid := range uicompletion.RequiredPermissions() {
		if slices.Contains(domain, pid) {
			t.Errorf("the domain role duplicates the framework permission %q; assign cfgai.RoleAssistantUser instead", pid)
		}
	}
}
