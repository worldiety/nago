// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package theme

import (
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/presentation/ui"
)

var _ = enum.Variant[settings.GlobalSettings, Settings](
	enum.Rename[Settings]("nago.theme.settings"),
)

type Colors struct {
	Dark  ui.Colors
	Light ui.Colors
}

type Settings struct {
	_ any `title:"Theme" description:"Theme und Einstellungen der Corporate Identity."`

	PageLogoLight image.ID `json:"pageLogoLight" section:"Logos" label:"Seitenlogo Light-Mode"`
	PageLogoDark  image.ID `json:"pageLogoDark" section:"Logos" label:"Seitenlogo Dark-Mode"`
	AppIconLight  image.ID `json:"appIconLight" section:"Logos" label:"App Icon Light-Mode"`
	AppIconDark   image.ID `json:"appIconDark" section:"Logos" label:"App Icon Dark-Mode"`

	Impress                   string `json:"impress" section:"Rechtliches" label:"Impressum" supportingText:"Link zum Impressum. Dies muss entweder ein absoluter externer Link sein beginnend mit https:// oder eine interne Navigationsroute wie page/impressum."`
	PrivacyPolicy             string `json:"privacyPolicy" section:"Rechtliches" label:"Datenschutz" supportingText:"Link zur Datenschutzerklärung. Dies muss entweder ein absoluter externer Link sein beginnend mit https:// oder eine interne Navigationsroute wie page/datenschutz."`
	GeneralTermsAndConditions string `json:"generalTermsAndConditions" section:"Rechtliches" label:"AGB" supportingText:"Link zur AGB. Dies muss entweder ein absoluter externer Link sein beginnend mit https:// oder eine interne Navigationsroute wie page/agb."`
	TermsOfUse                string `json:"termsOfUse" section:"Rechtliches" label:"Nutzungsbedingungen" supportingText:"Link zu den Nutzungsbedingungen. Nutzungsbedingungen regeln Handlungspflichten ohne Vertragsbeziehung, im Gegensatz zur AGB. Dies muss entweder ein absoluter externer Link sein beginnend mit https:// oder eine interne Navigationsroute wie page/nutzungsbedingungen."`
	ProviderName              string `json:"providerName" section:"Rechtliches" label:"Anbieter" supportingText:"Name des Anbieters der rechtlich für die Inhalte verantwortlich ist."`

	Slogan string `json:"slogan" section:"Sonstiges" label:"Slogan" supportingText:"Slogan oder Mission des Anbieters."`

	Colors Colors `json:"colors"` // TODO form.Auto cannot render that today, also this is not wanted from the designers perspective and it affects the global applications
}

func (s Settings) GlobalSettings() bool {
	return true
}
