// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uitheme

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

// colorField describes a field of [ui.Colors] for the advanced mode, by its json name.
type colorField struct {
	key   string
	label i18n.StrHnd
	desc  i18n.StrHnd
}

type colorGroup struct {
	title  i18n.StrHnd
	fields []colorField
}

func field(key, en, de, enDesc, deDesc string) colorField {
	return colorField{
		key:   key,
		label: i18n.MustString(i18n.Key("nago.theme.color."+key), i18n.Values{language.English: en, language.German: de}),
		desc:  i18n.MustString(i18n.Key("nago.theme.color."+key+"_desc"), i18n.Values{language.English: enDesc, language.German: deDesc}),
	}
}

func group(key, en, de string, fields ...colorField) colorGroup {
	return colorGroup{
		title:  i18n.MustString(i18n.Key("nago.theme.group."+key), i18n.Values{language.English: en, language.German: de}),
		fields: fields,
	}
}

// colorGroups lists the calculated colors of the advanced mode. The base colors M0, I0 and A0 are edited in the
// simple mode, thus they are not repeated here.
var colorGroups = []colorGroup{
	group("surfaces", "Surfaces and text", "Flächen und Text",
		field("M1", "Background", "Hintergrund", "Background of the page.", "Hintergrund der Seite."),
		field("M2", "Container", "Container", "Background of a first level container.", "Hintergrund eines Containers der ersten Ebene."),
		field("M3", "Card bottom", "Kartenfuß", "Bottom of a card.", "Unterer Bereich einer Karte."),
		field("M4", "Card body", "Kartenkörper", "Body of a card.", "Inhalt einer Karte."),
		field("M5", "Lines", "Linien", "Lines, dots and borders.", "Linien, Punkte und Rahmen."),
		field("M6", "Hovered", "Hover", "Hovered containers.", "Container unter dem Mauszeiger."),
		field("M7", "Muted text", "Gedämpfter Text", "Secondary text and muted icons.", "Nebensächlicher Text und gedämpfte Icons."),
		field("M8", "Text", "Text", "Normal text and icons.", "Normaler Text und Icons."),
		field("M9", "Card top", "Kartenkopf", "Top of a card.", "Oberer Bereich einer Karte."),
	),
	group("accents", "Accents", "Akzente",
		field("A1", "Accent", "Akzent", "Progress bars, charts, accented headlines and borders.", "Fortschrittsbalken, Diagramme, betonte Überschriften und Rahmen."),
		field("A2", "Accent 2", "Akzent 2", "Further accents.", "Weitere Akzente."),
	),
	group("interactive", "Interaction", "Interaktion",
		field("I1", "Buttons", "Knöpfe", "Background of buttons.", "Hintergrund von Knöpfen."),
		field("PBT", "Text on primary buttons", "Text auf Primärknöpfen", "Must contrast with the interactive color.", "Muss sich von der Interaktionsfarbe abheben."),
	),
	group("semantic", "Semantics", "Semantik",
		field("SE0", "Error", "Fehler", "Negative or destructive intentions.", "Negative oder zerstörerische Absichten."),
		field("SW0", "Warning", "Warnung", "Critical conditions.", "Kritische Zustände."),
		field("SG0", "Success", "Erfolg", "Positive conditions and confirmations.", "Positive Zustände und Bestätigungen."),
		field("SV0", "Information", "Information", "Something which has just changed.", "Etwas, das sich gerade geändert hat."),
		field("SI0", "Disabled", "Deaktiviert", "Disabled inputs.", "Deaktivierte Eingaben."),
		field("ST0", "Disabled text", "Deaktivierter Text", "Text on disabled inputs.", "Text auf deaktivierten Eingaben."),
	),
	group("banners", "Banners", "Banner",
		field("CBEB", "Error background", "Fehler Hintergrund", "Background of an error banner.", "Hintergrund eines Fehlerbanners."),
		field("CBET", "Error text", "Fehler Text", "Text of an error banner.", "Text eines Fehlerbanners."),
		field("CBIB", "Info background", "Info Hintergrund", "Background of an info banner.", "Hintergrund eines Infobanners."),
		field("CBIT", "Info text", "Info Text", "Text of an info banner.", "Text eines Infobanners."),
		field("CBWB", "Warning background", "Warnung Hintergrund", "Background of a warning banner.", "Hintergrund eines Warnbanners."),
		field("CBWT", "Warning text", "Warnung Text", "Text of a warning banner.", "Text eines Warnbanners."),
		field("CBSB", "Success background", "Erfolg Hintergrund", "Background of a success banner.", "Hintergrund eines Erfolgsbanners."),
		field("CBST", "Success text", "Erfolg Text", "Text of a success banner.", "Text eines Erfolgsbanners."),
	),
}
