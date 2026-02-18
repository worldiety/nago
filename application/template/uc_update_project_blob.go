// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"context"
	"io"
	"os"
	"sync"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
)

func NewUpdateProjectBlob(mutex *sync.Mutex, files blob.Store, repo Repository) UpdateProjectBlob {
	return func(subject auth.Subject, pid ID, filename string, value io.Reader) error {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(pid), PermUpdateProjectBlob); err != nil {
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
		for _, f := range prj.Files {
			if f.Filename == filename {
				wr, err := files.NewWriter(context.Background(), f.Blob)
				if err != nil {
					return err
				}

				defer wr.Close()

				_, err = io.Copy(wr, value)
				if err != nil {
					return err
				}

				return nil
			}
		}

		return os.ErrNotExist
	}
}
