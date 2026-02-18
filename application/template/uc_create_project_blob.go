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
	"time"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
)

func NewCreateProjectBlob(mutex *sync.Mutex, files blob.Store, repo Repository) CreateProjectBlob {
	return func(subject auth.Subject, pid ID, filename string, value io.Reader) error {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(pid), PermCreateProjectBlob); err != nil {
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

		newBlobId := data.RandIdent[string]()
		prj := optPrj.Unwrap()
		for _, f := range prj.Files {
			if f.Filename == filename {
				return os.ErrExist
			}

			if f.Blob == newBlobId {
				// check potential collision
				return os.ErrExist
			}
		}

		wr, err := files.NewWriter(context.Background(), newBlobId)
		if err != nil {
			return err
		}

		defer wr.Close()

		_, err = io.Copy(wr, value)
		if err != nil {
			return err
		}

		prj.Files = append(prj.Files, File{
			Filename: filename,
			Blob:     newBlobId,
			LastMod:  time.Now(),
		})

		return repo.Save(prj)
	}
}
