// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package library is a bounded context: it says what the system does, and nothing about how it is reached.
//
// It knows nothing about the user interface in ui/ and nothing about the AI tools in ai/ - both are ways of
// reaching this context, and the dependency only ever points inwards. That is what lets the assistant be
// built from these use cases without a single adapter: a use case already has the shape a tool needs, and it
// already audits the acting subject.
package library

// BookID identifies a book in the library.
type BookID string

// Book is a title the library owns, in some number of copies.
//
// This aggregate is handed to the model as it is - there is no separate DTO next to it. The `desc` tags are
// what makes that work: they are the documentation the model reads, exactly as `label` tags are what
// form.Auto reads. A second type that had to be kept in sync with this one would buy nothing and would
// eventually drift.
//
// Write a DTO when the model needs something this type does not have: a value formatted for a human, an
// aggregate computed across records, or a subset that deliberately hides a field. Not merely to rename
// things.
type Book struct {
	ID     BookID `json:"id" desc:"stable identifier, used when lending or returning"`
	Title  string `json:"title"`
	Author string `json:"author"`
	// Copies is how many the library owns in total.
	Copies int `json:"copies" desc:"total number of copies owned, lent out ones included"`
	// LentTo lists the borrowers currently holding a copy.
	//
	// The number of free copies is Copies minus the length of this list. It is not a field of its own on
	// purpose: a derived value that is also stored is a value that can disagree with itself.
	LentTo []string `json:"lentTo,omitempty" desc:"names of the people currently holding a copy; a book is fully lent out when this has as many entries as there are copies"`
}

// Identity makes the book an aggregate root.
func (b Book) Identity() BookID { return b.ID }

// WithIdentity is required by the repository.
func (b Book) WithIdentity(id BookID) Book {
	b.ID = id
	return b
}

// String is what a picker displays.
func (b Book) String() string { return b.Title }

// Available is how many copies can still be lent out.
func (b Book) Available() int {
	return b.Copies - len(b.LentTo)
}

// BookFilter narrows a listing.
//
// Every field is optional: a filter narrows a listing, it does not command one. The `optional` tags say so to
// a model that is handed this type as a tool input - without them the schema would demand all three on every
// call, and the model would dutifully invent them.
//
// The `desc` tags are not decoration either: they are the documentation the model reads to decide how to call
// this. Write them for a colleague who has never seen the application.
type BookFilter struct {
	Query         string `json:"query" optional:"true" desc:"case-insensitive substring matched against title and author"`
	OnlyAvailable bool   `json:"onlyAvailable" optional:"true" desc:"when true, only books with at least one free copy are returned"`
	Borrower      string `json:"borrower" optional:"true" desc:"only books currently lent to this borrower"`
}

// LendRequest asks for one copy of a book.
//
// Both fields are mandatory and neither carries `optional`, so the schema tells the model as much before it
// ever calls. That is the difference to [BookFilter]: a filter narrows, a request commands.
type LendRequest struct {
	Book     BookID `json:"book" desc:"the id of the book, as returned by list_books"`
	Borrower string `json:"borrower" desc:"the name of the person receiving the copy"`
}

// LendResult reports in words what happened, because that is what both a human and a model need back.
type LendResult struct {
	Summary   string `json:"summary"`
	Available int    `json:"available" desc:"copies still available after this operation"`
}
