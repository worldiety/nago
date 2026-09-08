// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package cfglibrary wires the library context into a running application.
//
// It is the layer where storage, screens and domain meet, and the only one that is allowed to. The context
// itself knows only its repository interface, so swapping the store never touches a use case, and it knows
// nothing about its own screens.
package cfglibrary

import (
	"fmt"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/app/library"
	uilibrary "go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/app/library/ui"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
)

// RoleLibrarian is what an operator assigns to give somebody the library.
const RoleLibrarian role.ID = "tutorial.library.librarian"

// Module is what the rest of the application depends on.
type Module struct {
	UseCases library.UseCases
	Pages    uilibrary.Pages
}

// Enable builds the library context, seeds it, ships its role and registers its screens.
func Enable(cfg *application.Configurator) (Module, error) {
	store, err := cfg.EntityStore("tutorial.library.book")
	if err != nil {
		return Module{}, fmt.Errorf("cannot open book store: %w", err)
	}

	repo := library.BookRepository(json.NewSloppyJSONRepository[library.Book, library.BookID](store))
	if err := library.Seed(repo); err != nil {
		return Module{}, fmt.Errorf("cannot seed the shelf: %w", err)
	}

	uc := library.NewUseCases(repo)

	if err := declareLibrarianRole(cfg); err != nil {
		return Module{}, err
	}

	pages := uilibrary.Pages{Books: "."}

	// The purpose is registered with the route rather than with a menu entry, because the registration is the
	// only declaration every screen has: not every screen appears in a menu, and menus are rearranged freely.
	// uicompletion.WindowContext reads it to tell the model where the user is standing.
	cfg.RootViewWithDecoration(pages.Books, func(wnd core.Window) core.View {
		return uilibrary.PageBooks(wnd, uc)
	}, application.Purpose(
		"Den Bestand der Bibliothek durchsehen: welche Titel es gibt, wie viele Exemplare frei sind und wer welches ausgeliehen hat."))

	return Module{UseCases: uc, Pages: pages}, nil
}

// declareLibrarianRole ships the application's own authorization as one assignable role.
//
// Two things are worth noticing here.
//
// First, the assistant permissions are deliberately NOT in this list. cfgai declares its own system role
// (cfgai.RoleAssistantUser) holding exactly the three framework permissions the chat needs, and assigning
// that role is what makes the button appear. Keeping them apart means an operator can decide per user whether
// they get the assistant, without touching what the user may do in the application - and the assistant,
// bounded by the acting subject, still cannot exceed the library permissions granted here.
//
// Second, this is a system role, so it cannot be deleted and its permission set cannot be edited away in the
// admin UI. Both would be undone on the next start anyway; refusing outright is the honest version.
func declareLibrarianRole(cfg *application.Configurator) error {
	return cfg.DeclareSystemRole(role.Role{
		ID:          RoleLibrarian,
		Name:        "Bibliothekar",
		Description: "Darf den Bestand einsehen sowie Exemplare ausleihen und zurücknehmen.",
	}, LibrarianPermissions()...)
}

// LibrarianPermissions is spelled out as its own function so a test can check it without reaching into the
// declaration.
func LibrarianPermissions() []permission.ID {
	return []permission.ID{
		library.PermFindAllBooks,
		library.PermLendBook,
		library.PermReturnBook,
	}
}
