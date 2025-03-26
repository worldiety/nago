// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"context"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	fs2 "go.wdy.de/nago/pkg/blob/fs"
	"io"
	"io/fs"
	"sync"
	"time"
)

func NewEnsureBuildIn(mutex *sync.Mutex, repository Repository, blobs blob.Store) EnsureBuildIn {
	return func(subject auth.Subject, project NewProjectData, force bool) error {
		if err := subject.Audit(PermEnsureBuildIn); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if project.ID == "" {
			return fmt.Errorf("buildin project id can't be empty")
		}

		optPrj, err := repository.FindByID(project.ID)
		if err != nil {
			return fmt.Errorf("project find by id err: %v", err)
		}

		ctx := context.Background()

		if optPrj.IsNone() || force {
			var prj Project
			prj.ID = project.ID
			prj.Name = project.Name
			prj.Type = project.ExecType
			prj.Description = project.Description
			prj.Tags = project.Tags
			err := fs.WalkDir(project.Files, ".", func(path string, d fs.DirEntry, err error) error {
				if d.Type().IsRegular() {
					hash, err := fs2.Sha512_224(project.Files, path)
					if err != nil {
						return fmt.Errorf("project hash err: %v", err)
					}

					blobName := hex.EncodeToString(hash[:])
					fileExists, err := blobs.Exists(ctx, blobName)
					if err != nil {
						return fmt.Errorf("cannot check if blob exists: %v", err)
					}

					prj.Files = append(prj.Files, File{
						Filename: path,
						Blob:     blobName,
						LastMod:  time.Now(),
					})

					if fileExists {
						// ignore, already saved, do not trigger hidden writes
						return nil
					}

					r, err := project.Files.Open(path)
					if err != nil {
						return fmt.Errorf("project open err: %v", err)
					}

					defer r.Close()

					ctx, cancel := context.WithCancel(ctx)
					w, err := blobs.NewWriter(ctx, blobName)
					if err != nil {
						cancel()
						return fmt.Errorf("blobs open err: %v", err)
					}

					if _, err := io.Copy(w, r); err != nil {
						cancel() // cancel before close, thus data is discard
						_ = w.Close()
						return fmt.Errorf("blobs copy err: %v", err)
					}

					// commit blob
					if err := w.Close(); err != nil {
						cancel()
						return fmt.Errorf("blobs close err: %v", err)
					}

					cancel() // cancel after close, thus data is committed

					return nil
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf("cannot import project files: %v", err)
			}

			if err := repository.Save(prj); err != nil {
				return fmt.Errorf("project save err: %v", err)
			}
		}

		return nil
	}
}
