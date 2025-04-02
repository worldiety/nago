// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package inspector

import "go.wdy.de/nago/application/permission"

var (
	PermDataInspector = permission.Declare[FindAll]("nago.data.inspector", "Alle Repositories untersuchen", "Träger dieser Berechtigung können alle Daten anzeigen, löschen und bearbeiten. Diese Funktion ist nur für Wartungszwecke gedacht.")
)
