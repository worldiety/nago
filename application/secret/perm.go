// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import "go.wdy.de/nago/application/permission"

var (
	PermFindMySecrets        = permission.Declare[FindMySecrets]("nago.secret.find_my_secrets", "Meine Secrets anzeigen", "Träger dieser Berechtigung können die Secrets anzeigen, die sie besitzen oder die in eine ihrer Gruppen geteilt wurden.")
	PermCreateSecret         = permission.Declare[CreateSecret]("nago.secret.create", "Ein privates Secret erstellen", "Träger dieser Berechtigung können private Secrets erstellen.")
	PermUpdateMySecretGroups = permission.Declare[UpdateMySecretGroups]("nago.secret.groups.update", "Gruppen Secrets zuweisen", "Träger dieser Berechtigung können Secrets, auf die sie Zugriff haben, Gruppen zuweisen, in denen sie ebenfalls Mitglied sind. Entfernt werden können nur Gruppen, in denen sie selbst Mitglied sind.")
	PermUpdateMySecretOwners = permission.Declare[UpdateMySecretOwners]("nago.secret.owners.update", "Besitzer Secrets zuweisen", "Träger dieser Berechtigung können bei Secrets, auf die sie Zugriff haben, die Besitzer ändern.")
	PermUpdateMyCredentials  = permission.Declare[UpdateMyCredentials]("nago.secret.credentials.update", "Meine Secret Credentials aktualisieren", "Träger dieser Berechtigung können die Credentials eines Secrets aktualisieren, das ihnen gehört oder das in eine ihrer Gruppen geteilt wurde.")
	PermDeleteMySecretByID   = permission.Declare[DeleteMySecretByID]("nago.secret.delete", "Mein Secret löschen", "Träger dieser Berechtigung können Secrets entfernen, die sie besitzen oder die in eine ihrer Gruppen geteilt wurden.")
)
