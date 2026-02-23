// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"slices"
	"sync"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
)

func NewDeleteProjectBlob(mutex *sync.Mutex, files blob.Store, repo Repository) DeleteProjectBlob {
	return func(subject auth.Subject, pid ID, filename string) error {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(pid), PermDeleteProjectBlob); err != nil {
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
