// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/auth"
)

// LendBook hands one copy of a book to a person.
type LendBook func(subject auth.Subject, req LendRequest) (LendResult, error)

// NewLendBook builds the lending use case.
func NewLendBook(repo BookRepository) LendBook {
	return func(subject auth.Subject, req LendRequest) (LendResult, error) {
		if err := subject.Audit(PermLendBook); err != nil {
			return LendResult{}, err
		}

		borrower := strings.TrimSpace(req.Borrower)
		if borrower == "" {
			return LendResult{}, fmt.Errorf("ohne Namen kann kein Exemplar herausgegeben werden")
		}

		optBook, err := repo.FindByID(req.Book)
		if err != nil {
			return LendResult{}, err
		}

		if optBook.IsNone() {
			return LendResult{}, fmt.Errorf("kein Buch mit der Kennung %q", req.Book)
		}

		book := optBook.Unwrap()
		if book.Available() <= 0 {
			return LendResult{}, fmt.Errorf("von %q ist derzeit kein Exemplar frei", book.Title)
		}

		book.LentTo = append(book.LentTo, borrower)
		if err := repo.Save(book); err != nil {
			return LendResult{}, err
		}

		// The result says what happened in words. A caller that is a model needs that as much as a human
		// does: it has to report back, and it must not have to reconstruct the outcome from the arguments it
		// sent.
		return LendResult{
			Summary:   fmt.Sprintf("%q an %s ausgeliehen.", book.Title, borrower),
			Available: book.Available(),
		}, nil
	}
}
