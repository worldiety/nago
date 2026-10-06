// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package data_test

import (
	"testing"

	"go.wdy.de/nago/pkg/data"
	datamem "go.wdy.de/nago/pkg/data/mem"
)

type entity struct {
	ID string
}

func (e entity) Identity() string {
	return e.ID
}

// Closing an observer removes exactly that observer, also after another one has been added.
func TestNotifyRepositoryClosesTheRightObserver(t *testing.T) {
	repo := data.NewNotifyRepository[entity, string](nil, &datamem.Repository[entity, string]{})

	var first, second, deletedFirst, deletedSecond int
	closeFirst := repo.AddSavedObserver(func(data.Repository[entity, string], data.Saved[entity, string]) error {
		first++
		return nil
	})
	repo.AddSavedObserver(func(data.Repository[entity, string], data.Saved[entity, string]) error {
		second++
		return nil
	})

	closeDeletedFirst := repo.AddDeletedObserver(func(data.Repository[entity, string], data.Deleted[string]) error {
		deletedFirst++
		return nil
	})
	repo.AddDeletedObserver(func(data.Repository[entity, string], data.Deleted[string]) error {
		deletedSecond++
		return nil
	})

	closeFirst()
	closeDeletedFirst()

	if err := repo.Save(entity{ID: "a"}); err != nil {
		t.Fatal(err)
	}

	if err := repo.DeleteByID("a"); err != nil {
		t.Fatal(err)
	}

	if first != 0 || second != 1 {
		t.Fatalf("the closed saved observer got %d calls, the other %d", first, second)
	}

	if deletedFirst != 0 || deletedSecond != 1 {
		t.Fatalf("the closed deleted observer got %d calls, the other %d", deletedFirst, deletedSecond)
	}
}
