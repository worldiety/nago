// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndbinspector

import "go.wdy.de/nago/application/permission"

// Inspect is the marker use-case type for the ndb inspector permission. The
// permission framework requires a named func type.
type Inspect func()

var (
	// PermNDBInspector grants full read and destructive (knife-tool) access to
	// the ndb inspector: browsing message streams and time series, and deleting
	// or repairing them. It is intended for maintenance only.
	PermNDBInspector = permission.Declare[Inspect](
		"nago.ndb.inspector",
		"ndb Datenbanken untersuchen",
		"Träger dieser Berechtigung können ndb Message- und Timeseries-Datenbanken einsehen, einzelne Einträge und ganze Streams löschen sowie Reparaturwerkzeuge ausführen. Nur für Wartungszwecke.",
	)
)
