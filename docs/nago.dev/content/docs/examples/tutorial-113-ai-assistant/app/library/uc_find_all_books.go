// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

import (
	"iter"
	"strings"

	"go.wdy.de/nago/auth"
)

// FindAllBooks reads the stock, narrowed by a filter.
//
// Reading is a promise too - that someone may see a thing - so a query is no different from any other use
// case: its own file, its own constructor, its own permission, its own place in the bundle.
type FindAllBooks func(subject auth.Subject, filter BookFilter) iter.Seq2[Book, error]

// NewFindAllBooks builds the listing query over the book repository.
func NewFindAllBooks(repo BookRepository) FindAllBooks {
	return func(subject auth.Subject, filter BookFilter) iter.Seq2[Book, error] {
		return func(yield func(Book, error) bool) {
			if err := subject.Audit(PermFindAllBooks); err != nil {
				yield(Book{}, err)
				return
			}

			for book, err := range repo.All() {
				if err != nil {
					yield(Book{}, err)
					return
				}

				if !filter.Matches(book) {
					continue
				}

				if !yield(book, nil) {
					return
				}
			}
		}
	}
}

// Matches reports whether a book falls into the filter.
func (f BookFilter) Matches(b Book) bool {
	if f.OnlyAvailable && b.Available() <= 0 {
		return false
	}

	if q := strings.ToLower(strings.TrimSpace(f.Query)); q != "" {
		if !strings.Contains(strings.ToLower(b.Title), q) && !strings.Contains(strings.ToLower(b.Author), q) {
			return false
		}
	}

	if f.Borrower != "" {
		if !b.isLentTo(f.Borrower) {
			return false
		}
	}

	return true
}

// isLentTo reports whether this person currently holds a copy. The comparison ignores case because the name
// is typed by a human - or, increasingly, produced by one from a sentence.
func (b Book) isLentTo(borrower string) bool {
	for _, who := range b.LentTo {
		if strings.EqualFold(who, borrower) {
			return true
		}
	}

	return false
}
