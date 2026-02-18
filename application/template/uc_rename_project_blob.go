// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"os"
	"sync"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
)

func NewRenameProjectBlob(mutex *sync.Mutex, files blob.Store, repo Repository) RenameProjectBlob {
	return func(subject auth.Subject, pid ID, filename string, newFilename string) error {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(pid), PermRenameProjectBlob); err != nil {
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
		foundOld := false
		foundNew := false
		fileIdx := -1
		for idx, f := range prj.Files {
			if f.Filename == filename {
				foundOld = true
				fileIdx = idx
			}

			if f.Filename == newFilename {
				foundNew = true
			}
		}

		if !foundOld {
			return os.ErrNotExist
		}

		if foundNew {
			return os.ErrExist
		}

		prj.Files[fileIdx].Filename = newFilename

		return repo.Save(prj)
	}
}
