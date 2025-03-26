// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"context"
	"fmt"
	"sync"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
)

func NewDelete(mutex *sync.Mutex, files blob.Store, repo Repository) Delete {
	return func(subject auth.Subject, pid ID) error {
		if err := subject.AuditResource(repo.Name(), string(pid), PermDelete); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optPrj, err := repo.FindByID(pid)
		if err != nil {
			return err
		}

		if optPrj.IsNone() {
			return nil
		}

		prj := optPrj.Unwrap()
		ctx := context.Background()

		for _, file := range prj.Files {
			if err := files.Delete(ctx, file.Blob); err != nil {
				return fmt.Errorf("cannot delete file %s: %w", file.Filename, err)
			}
		}

		return repo.DeleteByID(prj.ID)
	}
}
