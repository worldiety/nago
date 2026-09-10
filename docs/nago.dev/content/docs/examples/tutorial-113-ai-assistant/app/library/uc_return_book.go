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

// ReturnBook takes a lent copy back from a person.
type ReturnBook func(subject auth.Subject, req LendRequest) (LendResult, error)

// NewReturnBook builds the return use case.
func NewReturnBook(repo BookRepository) ReturnBook {
	return func(subject auth.Subject, req LendRequest) (LendResult, error) {
		if err := subject.Audit(PermReturnBook); err != nil {
			return LendResult{}, err
		}

		optBook, err := repo.FindByID(req.Book)
		if err != nil {
			return LendResult{}, err
		}

		if optBook.IsNone() {
			return LendResult{}, fmt.Errorf("kein Buch mit der Kennung %q", req.Book)
		}

		book := optBook.Unwrap()

		idx := -1
		for i, who := range book.LentTo {
			if strings.EqualFold(who, req.Borrower) {
				idx = i
				break
			}
		}

		if idx < 0 {
			return LendResult{}, fmt.Errorf("%s hat kein Exemplar von %q", req.Borrower, book.Title)
		}

		book.LentTo = append(book.LentTo[:idx], book.LentTo[idx+1:]...)
		if err := repo.Save(book); err != nil {
			return LendResult{}, err
		}

		return LendResult{
			Summary:   fmt.Sprintf("%q von %s zurückerhalten.", book.Title, req.Borrower),
			Available: book.Available(),
		}, nil
	}
}
