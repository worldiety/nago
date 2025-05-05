// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/permission"
)

var (
	PermCreate                    = permission.Declare[Create]("nago.user.create", "Nutzer anlegen", "Träger dieser Berechtigung können neue Nutzer anlegen.")
	PermFindByID                  = permission.Declare[FindByID]("nago.user.find_by_id", "Einen Nutzer per ID finden", "Träger dieser Berechtigung können die Eigenschaften anderer Nutzer anzeigen.")
	PermFindByMail                = permission.Declare[FindByMail]("nago.user.find_by_mail", "Einen Nutzer per Mail finden", "Träger dieser Berechtigung können die Eigenschaften anderer Nutzer anzeigen.")
	PermFindAll                   = permission.Declare[FindAll]("nago.user.find_all", "Alle Nutzer finden", "Träger dieser Berechtigung können vorhandene Nutzer und ihre Eigenschaften anzeigen.")
	PermChangeOtherPassword       = permission.Declare[ChangeOtherPassword]("nago.user.change_other_password", "Kennwort ändern", "Träger dieser Berechtigung können vorhandene Nutzer und ihre Eigenschaften anzeigen.")
	PermDelete                    = permission.Declare[Delete]("nago.user.delete", "Nutzer Löschen", "Träger dieser Berechtigung können vorhandene Nutzer und ihre Eigenschaften anzeigen.")
	PermUpdateOtherContact        = permission.Declare[UpdateOtherContact]("nago.user.update_other_contact", "Kontaktdaten von Nutzern ändern", "Träger dieser Berechtigung können die Kontaktdaten vorhandener Nutzer aktualisieren.")
	PermUpdateOtherRoles          = permission.Declare[UpdateOtherRoles]("nago.user.update_other_roles", "Rollenmitgliedschaft von Nutzern ändern", "Träger dieser Berechtigung können die Rollenmitgliedschaften vorhandener Nutzer aktualisieren.")
	PermUpdateOtherPermissions    = permission.Declare[UpdateOtherPermissions]("nago.user.update_other_permissions", "Berechtigungen von Nutzern ändern", "Träger dieser Berechtigung können die individuellen Berechtigungen vorhandener Nutzer aktualisieren.")
	PermUpdateOtherLicenses       = permission.Declare[UpdateOtherLicenses]("nago.user.update_other_licenses", "Lizenzen von Nutzern ändern", "Träger dieser Berechtigung können die individuellen Lizenzen vorhandener Nutzer aktualisieren.")
	PermUpdateOtherGroups         = permission.Declare[UpdateOtherGroups]("nago.user.update_other_groups", "Gruppenzugehörigkeit von Nutzern ändern", "Träger dieser Berechtigung können die Kontaktdaten vorhandener Nutzer aktualisieren.")
	PermUpdateaccountStatus       = permission.Declare[UpdateAccountStatus]("nago.user.update_account_status", "Account Status von Nutzern ändern", "Träger dieser Berechtigung können Nutzer aktivieren oder deaktivieren.")
	PermCountAssignedUserLicense  = permission.Declare[CountAssignedUserLicense]("nago.user.count_assigned_user_license", "Anzahl Nutzerlizenzen ermitteln", "Träger dieser Berechtigung können die Menge einer zugewiesenen nutzerbasierten Lizenz ermitteln.")
	PermRevokeAssignedUserLicense = permission.Declare[RevokeAssignedUserLicense]("nago.user.revoke_assigned_user_license", "Anzahl Nutzerlizenzen entfernen", "Träger dieser Berechtigung können eine Menge an zugewiesenen nutzerbasierten Lizenz von Nutzern anonym entfernen.")
	PermAssignUserLicense         = permission.Declare[AssignUserLicense]("nago.user.assign_user_license", "Einem Nutzer eine Lizenz zuweisen", "Träger dieser Berechtigung können eine beliebige Lizenz einem beliebigen Nutzer zuweisen.")

	PermAddResourcePermissions    = permission.Declare[AddResourcePermissions]("nago.user.resource.addperm", "Eine Resourcen-Berechtigung einem Nutzer zuweisen", "Träger dieser Berechtigung können einem Nutzer eine Resourcen-orientierte Berechtigung zuweisen.")
	PermRemoveResourcePermissions = permission.Declare[RemoveResourcePermissions]("nago.user.resource.removeperm", "Eine Resourcen-Berechtigung entfernen", "Träger dieser Berechtigung können die Resourcen-orientierte Berechtigungen eines Nutzers entfernen.")
	PermListResourcePermissions   = permission.Declare[ListResourcePermissions]("nago.user.resource.listperm", "Resourcen-Berechtigung auflisten", "Träger dieser Berechtigung können die Resourcen-orientierte Berechtigungen eines Nutzers auflisten.")
)
