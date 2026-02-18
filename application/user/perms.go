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
	PermCreate                 = permission.Declare[Create]("nago.user.create", "Nutzer anlegen", "Träger dieser Berechtigung können neue Nutzer anlegen.")
	PermFindByID               = permission.Declare[FindByID]("nago.user.find_by_id", "Einen Nutzer per ID finden", "Träger dieser Berechtigung können die Eigenschaften anderer Nutzer anzeigen.")
	PermFindByMail             = permission.Declare[FindByMail]("nago.user.find_by_mail", "Einen Nutzer per Mail finden", "Träger dieser Berechtigung können die Eigenschaften anderer Nutzer anzeigen.")
	PermFindAll                = permission.Declare[FindAll]("nago.user.find_all", "Alle Nutzer finden", "Träger dieser Berechtigung können vorhandene Nutzer und ihre Eigenschaften anzeigen.")
	PermChangeOtherPassword    = permission.Declare[ChangeOtherPassword]("nago.user.change_other_password", "Kennwort ändern", "Träger dieser Berechtigung können vorhandene Nutzer und ihre Eigenschaften anzeigen.")
	PermDelete                 = permission.Declare[Delete]("nago.user.delete", "Nutzer Löschen", "Träger dieser Berechtigung können vorhandene Nutzer und ihre Eigenschaften anzeigen.")
	PermUpdateOtherContact     = permission.Declare[UpdateOtherContact]("nago.user.update_other_contact", "Kontaktdaten von Nutzern ändern", "Träger dieser Berechtigung können die Kontaktdaten vorhandener Nutzer aktualisieren.")
	PermUpdateOtherRoles       = permission.Declare[UpdateOtherRoles]("nago.user.update_other_roles", "Rollenmitgliedschaft von Nutzern ändern", "Träger dieser Berechtigung können die Rollenmitgliedschaften vorhandener Nutzer aktualisieren.")
	PermUpdateOtherPermissions = permission.Declare[UpdateOtherPermissions]("nago.user.update_other_permissions", "Berechtigungen von Nutzern ändern", "Träger dieser Berechtigung können die individuellen Berechtigungen vorhandener Nutzer aktualisieren.")
	PermUpdateOtherGroups      = permission.Declare[UpdateOtherGroups]("nago.user.update_other_groups", "Gruppenzugehörigkeit von Nutzern ändern", "Träger dieser Berechtigung können die Kontaktdaten vorhandener Nutzer aktualisieren.")
	PermUpdateaccountStatus    = permission.Declare[UpdateAccountStatus]("nago.user.update_account_status", "Account Status von Nutzern ändern", "Träger dieser Berechtigung können Nutzer aktivieren oder deaktivieren.")
	PermExportUsers            = permission.Declare[ExportUsers]("nago.user.export_users", "Ausgewählte Nutzer exportieren", "Träger dieser Berechtigung können die Kontaktdaten beliebiger Nutzer exportieren.")

	// deprecated use rebac api
	PermAddResourcePermissions = permission.Declare[AddResourcePermissions]("nago.user.resource.addperm", "Eine Resourcen-Berechtigung einem Nutzer zuweisen", "Träger dieser Berechtigung können einem Nutzer eine Resourcen-orientierte Berechtigung zuweisen.")
	// deprecated use rebac api
	PermRemoveResourcePermissions = permission.Declare[RemoveResourcePermissions]("nago.user.resource.removeperm", "Eine Resourcen-Berechtigung entfernen", "Träger dieser Berechtigung können die Resourcen-orientierte Berechtigungen eines Nutzers entfernen.")
	// deprecated use rebac api
	PermListResourcePermissions = permission.Declare[ListResourcePermissions]("nago.user.resource.listperm", "Resourcen-Berechtigung auflisten", "Träger dieser Berechtigung können die Resourcen-orientierte Berechtigungen eines Nutzers auflisten.")

	// deprecated use rebac api
	PermGrantPermissions = permission.Declare[GrantPermissions]("nago.grant.grant", "Grant permissions to others", "A user with that permission assigned can grant permissions to other users.")
	// deprecated use rebac api
	PermListGrantedUsers = permission.Declare[ListGrantedUsers]("nago.grant.listgranted", "List granted users for resource", "A user with that permission assigned can list other users which have granted permissions on a specific resource.")
	// deprecated use rebac api
	PermListGrantedPermissions = permission.Declare[ListGrantedPermissions]("nago.grant.listgrants", "List permissions for a users resource", "A user with that permission assigned can list granted permissions for specific user and resource.")

	PermConsent = permission.Declare[Consent]("nago.user.consent_other", "Zustimmungen anderer Nutzer setzen", "Träger dieser Berechtigung können die Datenschutz, Nutzungsbedingungen oder sonstige Erlaubnisse in deren Namen zustimmen.")
)
