// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package libsync

import (
	"fmt"
	"os"
	"slices"
	"sync"

	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

func NewAddSource(mutex *sync.Mutex, repo Repository) AddSource {
	return func(subject auth.Subject, id library.ID, src Source) error {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(id), Permissions.Update); err != nil {
			return err
		}

		if !src.Drive.Valid && !src.Store.Valid {
			return fmt.Errorf("src has no valid origins")
		}

		if src.Store.Valid && src.Store.Name == "" {
			return fmt.Errorf("src store has no valid name")
		}

		if src.Drive.Valid && src.Drive.Root == "" {
			return fmt.Errorf("src drive has no valid root")
		}

		mutex.Lock()
		defer mutex.Unlock()

		optJob, err := repo.FindByID(id)
		if err != nil {
			return err
		}

		if optJob.IsNone() {
			return fmt.Errorf("job not found %s: %w", id, os.ErrNotExist)
		}

		job := optJob.Unwrap()
		job.Sources = slices.DeleteFunc(job.Sources, func(source Source) bool {
			return source == src // clean-up, e.g. purge duplicates etc.
		})

		job.Sources = append(job.Sources, src)

		return repo.Save(job)
	}
}
