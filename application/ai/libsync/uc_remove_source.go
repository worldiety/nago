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
	"go.wdy.de/nago/auth"
)

func NewRemoveSource(mutex *sync.Mutex, repo Repository) RemoveSource {
	return func(subject auth.Subject, id library.ID, src Source) error {
		if err := subject.AuditResource(repo.Name(), string(id), Permissions.Update); err != nil {
			return err
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
			return source == src
		})

		return repo.Save(job)
	}
}
