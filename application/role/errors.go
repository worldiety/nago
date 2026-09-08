// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package role

import (
	"go.wdy.de/nago/pkg/std"
)

// errSystemRoleProtected is returned when a caller tries to delete a system role or to replace its
// permissions. It is a business rule, not a permission problem: not even a super user may do it, because the
// permission set belongs to the feature that declared the role and is restored on the next start anyway.
func errSystemRoleProtected(id ID) error {
	return std.NewLocalizedError(
		"Systemrolle",
		"Die Rolle "+string(id)+" wird von der Anwendung verwaltet. Sie kann weder gelöscht noch in ihren Berechtigungen verändert werden. Mitglieder können weiterhin zugewiesen werden.",
	)
}
