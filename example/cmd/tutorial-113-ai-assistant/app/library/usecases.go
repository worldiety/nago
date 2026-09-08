// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

// UseCases bundles the capabilities of the library context.
//
// Callers depend on this bundle rather than on the internals, and [NewUseCases] is the single place where the
// repository is threaded through. The screens in ui/ and the tools in ai/ both take it, and neither can reach
// past it into the store.
type UseCases struct {
	FindAllBooks FindAllBooks
	LendBook     LendBook
	ReturnBook   ReturnBook
}

// NewUseCases wires the library use cases.
func NewUseCases(repo BookRepository) UseCases {
	return UseCases{
		FindAllBooks: NewFindAllBooks(repo),
		LendBook:     NewLendBook(repo),
		ReturnBook:   NewReturnBook(repo),
	}
}
