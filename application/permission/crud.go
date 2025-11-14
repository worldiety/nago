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

// DeclareDeleteAll creates a prototype permission stub. See also [DeclareCreate].
func DeclareDeleteAll[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("Delete all %s elements", entityName),
			language.German:  fmt.Sprintf("Alle %s Element löschen", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can delete all " + entityName + " elements.",
			language.German:  "Träger dieser Berechtigung können alle " + entityName + "-Elemente löschen.",
		},
	).String()}, 3)
}

// DeclareReloadAll creates a prototype permission stub. See also [DeclareCreate].
func DeclareReloadAll[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("Reload all %s elements", entityName),
			language.German:  fmt.Sprintf("Alle %s Element neu laden", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can reload all " + entityName + " elements.",
			language.German:  "Träger dieser Berechtigung können alle " + entityName + "-Elemente neu laden.",
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

// DeclareAppend creates a prototype permission stub. See also [DeclareCreate].
func DeclareAppend[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("Append %s elements", entityName),
			language.German:  fmt.Sprintf("%s Elemente anhängen", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can get " + entityName + " append elements.",
			language.German:  "Träger dieser Berechtigung können " + entityName + "-Elemente anhängen.",
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

// DeclareSend creates a prototype permission stub. See also [DeclareCreate].
func DeclareSend[UseCase any](id ID, entityName string) ID {
	return register[UseCase](Permission{ID: id, Name: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_name", id)),
		i18n.Values{
			language.English: fmt.Sprintf("Send %s element", entityName),
			language.German:  fmt.Sprintf("%s Element senden", entityName),
		},
	).String(), Description: i18n.MustString(
		i18n.Key(fmt.Sprintf("%s_perm_desc", id)),
		i18n.Values{
			language.English: "Holders of this authorisation can send " + entityName + " elements.",
			language.German:  "Träger dieser Berechtigung können " + entityName + "-Elemente senden.",
		},
	).String()}, 3)
}
