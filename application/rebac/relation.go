// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

// Relation defines the relation between two entities and should be one of the constants below, however,
// it can be entirely arbitrary.
type Relation string

// Name either returns the predefined name of the relation or a translation identifier.
func (r Relation) Name() string {
	if v, ok := relationNames[r]; ok {
		return v.String()
	}

	return string(r)
}

// Description either returns the predefined description or a translation identifier.
func (r Relation) Description() string {
	if v, ok := relationDescriptions[r]; ok {
		return v.String()
	}

	return ""
}

const (
	Owner   Relation = "owner"
	Writer  Relation = "writer"
	Deleter Relation = "deleter"
	// Member is has-member semantics.
	Member Relation = "member"
	Viewer Relation = "viewer"
	Parent Relation = "parent"
)

var relationNames = map[Relation]i18n.StrHnd{
	Owner:   i18n.MustString("nago.relation.owner.name", i18n.Values{language.German: "Besitzer", language.English: "Owner"}),
	Writer:  i18n.MustString("nago.relation.writer.name", i18n.Values{language.German: "Schreiber", language.English: "Writer"}),
	Deleter: i18n.MustString("nago.relation.deleter.name", i18n.Values{language.German: "Löschender", language.English: "Deleter"}),
	Member:  i18n.MustString("nago.relation.member.name", i18n.Values{language.German: "Mitglied", language.English: "Member"}),
	Viewer:  i18n.MustString("nago.relation.viewer.name", i18n.Values{language.German: "Viewer", language.English: "Viewer"}),
	Parent:  i18n.MustString("nago.relation.parent.name", i18n.Values{language.German: "Parent", language.English: "Parent"}),
}

var relationDescriptions = map[Relation]i18n.StrHnd{
	Owner:   i18n.MustString("nago.relation.owner.description", i18n.Values{language.German: "Das Source ist der Besitzer des zugeordneten Objektes.", language.English: "The subject is the owner of the associated object."}),
	Writer:  i18n.MustString("nago.relation.writer.description", i18n.Values{language.German: "Das Source kann das zugeordnete Objekt schreiben.", language.English: "The subject can write the associated object."}),
	Deleter: i18n.MustString("nago.relation.deleter.description", i18n.Values{language.German: "Das Source kann das zugeordnete Objekt löschen.", language.English: "The subject can delete the associated object."}),
	Member:  i18n.MustString("nago.relation.member.description", i18n.Values{language.German: "Das Source ist Mitglied des zugeordneten Objektes.", language.English: "The subject is a member of the associated object."}),
	Viewer:  i18n.MustString("nago.relation.viewer.description", i18n.Values{language.German: "Das Source kann das zugeordnete Objekt lesen.", language.English: "The subject can read the associated object."}),
	Parent:  i18n.MustString("nago.relation.parent.description", i18n.Values{language.German: "Das Source ist das Elternelement des zugeordneten Objekts.", language.English: "The subject is the parent of the associated object."}),
}
