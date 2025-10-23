// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package permission

import (
	"fmt"

	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

// DeclareCreate returns a permission using the given id and a default text using the entity name and a simple
// generic text in english and german. The entity name is not translated and should match the
// name of the thing from the ubiquitous domain specific language. This will never work perfectly, but it helps
// to start over trivial CRUD-like use cases.
func DeclareCreate[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("Create %s element", entityName),
			language.German:  fmt.Sprintf("%s Element erstellen", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can create " + entityName + " elements.",
			language.German:  "Träger dieser Berechtigung können " + entityName + "-Elemente erstellen.",
		},
	).String()}, 3)
}

// DeclareFindByID creates a prototype permission stub. See also [DeclareCreate].
func DeclareFindByID[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("View %s element", entityName),
			language.German:  fmt.Sprintf("%s Element anzeigen", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can view " + entityName + " elements if they know their identifier.",
			language.German:  "Träger dieser Berechtigung können " + entityName + "-Elemente ansehen, wenn sie das Kennzeichen kennen.",
		},
	).String()}, 3)
}

// DeclareFindByName creates a prototype permission stub. See also [DeclareCreate].
func DeclareFindByName[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("View %s element by name", entityName),
			language.German:  fmt.Sprintf("%s Element nach Namen anzeigen", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can view " + entityName + " elements if they know their name.",
			language.German:  "Träger dieser Berechtigung können " + entityName + "-Elemente ansehen, wenn sie den Namen kennen.",
		},
	).String()}, 3)
}

// DeclareDeleteByID creates a prototype permission stub. See also [DeclareCreate].
func DeclareDeleteByID[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("Delete %s element", entityName),
			language.German:  fmt.Sprintf("%s Element löschen", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can delete " + entityName + " elements if they know their identifier.",
			language.German:  "Träger dieser Berechtigung können " + entityName + "-Elemente löschen, wenn sie das Kennzeichen kennen.",
		},
	).String()}, 3)
}

// DeclareFindAll creates a prototype permission stub. See also [DeclareCreate].
func DeclareFindAll[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("Get %s elements", entityName),
			language.German:  fmt.Sprintf("%s Elemente abrufen", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can get " + entityName + " all elements.",
			language.German:  "Träger dieser Berechtigung können alle " + entityName + "-Elemente abrufen.",
		},
	).String()}, 3)
}

// DeclareFindAllIdentifiers creates a prototype permission stub. See also [DeclareCreate].
func DeclareFindAllIdentifiers[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("Get %s element identifiers", entityName),
			language.German:  fmt.Sprintf("%s Element-Kennzeichen abrufen", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can get " + entityName + " all identifiers.",
			language.German:  "Träger dieser Berechtigung können alle " + entityName + "-Kennzeichen abrufen.",
		},
	).String()}, 3)
}

// DeclareUpdate creates a prototype permission stub. See also [DeclareCreate].
func DeclareUpdate[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("Update %s element", entityName),
			language.German:  fmt.Sprintf("%s Element aktualisieren", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can update " + entityName + " elements if they know their identifier.",
			language.German:  "Träger dieser Berechtigung können " + entityName + "-Elemente aktualisieren, wenn sie das Kennzeichen kennen.",
		},
	).String()}, 3)
}

// DeclareSync creates a prototype permission stub. See also [DeclareCreate].
func DeclareSync[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("Synchronize %s element", entityName),
			language.German:  fmt.Sprintf("%s Element synchronisieren", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can synchronize " + entityName + " elements if they know their identifier.",
			language.German:  "Träger dieser Berechtigung können " + entityName + "-Elemente synchronisieren, wenn sie das Kennzeichen kennen.",
		},
	).String()}, 3)
}
