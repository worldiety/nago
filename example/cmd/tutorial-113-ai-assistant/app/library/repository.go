// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

import (
	"go.wdy.de/nago/pkg/data"
)

// BookRepository stores the books. The context declares what it needs and never how it is stored; which
// store this is, is decided once in cfg/.
type BookRepository = data.Repository[Book, BookID]

// Seed puts something on the shelf so the example has something to talk about. It does nothing when the
// shelf is not empty, so a restart does not duplicate the stock.
func Seed(repo BookRepository) error {
	count, err := repo.Count()
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	books := []Book{
		{ID: "b1", Title: "Der Prozess", Author: "Franz Kafka", Copies: 2},
		{ID: "b2", Title: "Die Verwandlung", Author: "Franz Kafka", Copies: 1},
		{ID: "b3", Title: "Effective Go", Author: "The Go Authors", Copies: 3},
		{ID: "b4", Title: "Der Steppenwolf", Author: "Hermann Hesse", Copies: 1, LentTo: []string{"Anna"}},
	}

	for _, b := range books {
		if err := repo.Save(b); err != nil {
			return err
		}
	}

	return nil
}
