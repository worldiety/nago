// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uievs

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var (
	StrElements         = i18n.MustString("nago.evs.elements", i18n.Values{language.English: "Elements", language.German: "Elemente"})
	StrDataManagement   = i18n.MustString("nago.evs.datamanagement", i18n.Values{language.English: "Domain events", language.German: "Ereignisse verwalten"})
	StrManageEntitiesX  = i18n.MustVarString("nago.evs.manage_x", i18n.Values{language.English: "Manage domain events of type {name}. Publish or audit {name} events.", language.German: "Verwalte Domänenevents vom Typ {name}. Veröffentliche oder auditiere {name}-Ereignisse."})
	StrCreateEvtX       = i18n.MustVarString("nago.evs.create_evt_x", i18n.Values{language.English: "Create {name} event", language.German: "Erstelle {name}-Ereignis"})
	StrCreateDisclaimer = i18n.MustString("nago.evs.create_disclaimer", i18n.Values{language.English: "Create a domain event manually. Use this only for debugging, testing or repair purposes. Usually, events are emitted and stored by command (use case) execution.", language.German: "Erstellen Sie ein Domain-Ereignis manuell. Verwenden Sie diese Option nur zu Debugging-, Test- oder Reparaturzwecken. Normalerweise werden Ereignisse durch die Ausführung von Befehlen (Anwendungsfällen) ausgelöst und gespeichert."})
	StrIndexManagement  = i18n.MustVarString("nago.evs.manage_idx_x", i18n.Values{language.English: "Manage indexed domain events of type {name}.", language.German: "Verwalte indizierte Domänenevents vom Typ {name}."})
)
