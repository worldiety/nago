// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import "go.wdy.de/nago/application/permission"

var (
	PermSendMail             = permission.Declare[SendMail]("nago.mail.send", "Mail Senden", "Träger dieser Berechtigung können Mails versenden.")
	PermInitDefaultTemplates = permission.Declare[SendMail]("nago.mail.init_default_templates", "Standard Templates setzen", "Träger dieser Berechtigung können die Standard Mail templates aktivieren.")

	// the identifiers of the following permissions are kept stable from the former generic crud implementation,
	// so that existing role assignments stay valid.

	PermOutgoingFindAll    = permission.Declare[FindOutgoingIDs]("nago.mail.outgoing.find_all", "Ausgehende Mails anzeigen", "Träger dieser Berechtigung können die Warteschlange der ausgehenden Mails einsehen.")
	PermOutgoingFindByID   = permission.Declare[FindOutgoingByID]("nago.mail.outgoing.find_by_id", "Ausgehende Mail anzeigen", "Träger dieser Berechtigung können eine ausgehende Mail inklusive Inhalt und Versandversuchen einsehen.")
	PermOutgoingDeleteByID = permission.Declare[DeleteOutgoingByID]("nago.mail.outgoing.delete_by_id", "Ausgehende Mail löschen", "Träger dieser Berechtigung können ausgehende Mails aus der Warteschlange löschen.")
	PermOutgoingUpdate     = permission.Declare[SaveMail]("nago.mail.outgoing.update", "Ausgehende Mail aktualisieren", "Träger dieser Berechtigung können ausgehende Mails der Warteschlange überschreiben.")
	PermOutgoingRetry      = permission.Declare[RetryOutgoing]("nago.mail.outgoing.retry", "Versand erneut versuchen", "Träger dieser Berechtigung können fehlgeschlagene Mails erneut in die Warteschlange stellen.")
	PermOutgoingResend     = permission.Declare[ResendOutgoing]("nago.mail.outgoing.resend", "Mail erneut senden", "Träger dieser Berechtigung können eine Kopie einer vorhandenen Mail erneut versenden.")
	PermStatistics         = permission.Declare[Statistics]("nago.mail.statistics", "Mail-Statistiken anzeigen", "Träger dieser Berechtigung können Versandstatistiken, Probleme und den Zustand der SMTP-Server einsehen.")
)
