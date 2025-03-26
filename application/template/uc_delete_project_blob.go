// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"slices"
	"sync"
)

func NewDeleteProjectBlob(mutex *sync.Mutex, files blob.Store, repo Repository) DeleteProjectBlob {
	return func(subject auth.Subject, pid ID, filename string) error {
		if err := subject.AuditResource(repo.Name(), string(pid), PermDeleteProjectBlob); err != nil {
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
		prj.Files = slices.DeleteFunc(prj.Files, func(e File) bool {
			return e.Filename == filename
		})
		
		return repo.Save(prj)
	}
}
