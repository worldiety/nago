// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import "go.wdy.de/nago/application/permission"

var (
	PermRegisterParser               = permission.Declare[RegisterParser]("nago.dataimport.parser.register", "Datenimport-Parser registrieren", "Träger dieser Berechtigung können neue Parser registrieren.")
	PermRegisterImporter             = permission.Declare[RegisterImporter]("nago.dataimport.importer.register", "Datenimporter registrieren", "Träger dieser Berechtigung können neue Importer registrieren.")
	PermParse                        = permission.Declare[Parse]("nago.dataimport.parse", "Datenimport parsen", "Träger dieser Berechtigung können Rohdaten als Entwürfe parsen.")
	PermImport                       = permission.Declare[Import]("nago.dataimport.import", "Datenimport durchführen", "Träger dieser Berechtigung können Daten aus dem Entwurfsbereich importieren.")
	PermFindImporters                = permission.Declare[FindImporters]("nago.dataimport.findimporter", "Datenimporter anzeigen", "Träger dieser Berechtigung können Importer anzeigen.")
	PermFindParsers                  = permission.Declare[FindParsers]("nago.dataimport.findparsers", "Datenimport-Parser anzeigen", "Träger dieser Berechtigung können Parser anzeigen.")
	PermFindStaging                  = permission.Declare[FindStagingsForImporter]("nago.dataimport.findstaging", "Datenimport-Entwürfe anzeigen", "Träger dieser Berechtigung können Import Entwürfe anzeigen.")
	PermCreateStaging                = permission.Declare[CreateStaging]("nago.dataimport.createstaging", "Datenimport-Entwürfe erstellen", "Träger dieser Berechtigung können Import Entwürfe erstellen.")
	PermDeleteStaging                = permission.Declare[DeleteStaging]("nago.dataimport.deletestaging", "Datenimport-Entwürfe löschen", "Träger dieser Berechtigung können Import Entwürfe löschen.")
	PermFilterEntries                = permission.Declare[FilterEntries]("nago.dataimport.filterentries", "Datenimport-Entwürfe Einträge suchen", "Träger dieser Berechtigung können Einträge der Import Entwürfe ansehen und suchen.")
	PermUpdateStagingTransformation  = permission.Declare[UpdateStagingTransformation]("nago.dataimport.updatestagingtransformation", "Transformations-Datenimport-Entwürfe aktualisieren", "Träger dieser Berechtigung können die Transformation von Import-Entwürfen aktualisieren.")
	PermFindEntryByID                = permission.Declare[FindEntryByID]("nago.dataimport.findentrybyid", "Datenimport Entwurfseintrag lesen", "Träger dieser Berechtigung können einen Entwurfseintrag lesen.")
	PermUpdateEntryConfirmation      = permission.Declare[UpdateEntryConfirmation]("nago.dataimport.entry.updateconfirmation", "Datenimport Entwurfseintrag bestätigen", "Träger dieser Berechtigung können einen Entwurfseintrag bestätigen.")
	PermUpdateEntryIgnored           = permission.Declare[UpdateEntryIgnored]("nago.dataimport.entry.updateignored", "Datenimport Entwurfseintrag ignorieren", "Träger dieser Berechtigung können einen Entwurfseintrag ignorieren.")
	PermUpdateEntryTransformed       = permission.Declare[UpdateEntryTransformed]("nago.dataimport.entry.updatetransformed", "Datenimport Entwurfseintrag Transformationsmodell aktualisieren", "Träger dieser Berechtigung können das manuelle Transformationergebnis eines Entwurfseintrag aktualisieren.")
	PermCalculateStagingReviewStatus = permission.Declare[CalculateStagingReviewStatus]("nago.dataimport.entry.calculatestagingstatus", "Datenimport Entwurf Status berechnen", "Träger dieser Berechtigung können für einen Entwurf den Status berechnen lassen.")
)
