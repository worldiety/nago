// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package billing

import "go.wdy.de/nago/application/permission"

var (
	PermAppLicenses  = permission.Declare[AppLicenses]("nago.billing.license.app", "Abgerechnete App-Lizenzen anzeigen", "Träger dieser Berechtigung können alle verfügbaren und gebuchten Anwendungslizenzen anzeigen.")
	PermUserLicenses = permission.Declare[UserLicenses]("nago.billing.license.user", "Abgerechnete User-Lizenzen anzeigen", "Träger dieser Berechtigung können alle verfügbaren und gebuchten Nutzerlizenzen anzeigen.")
)
