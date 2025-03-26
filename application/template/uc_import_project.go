// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"archive/zip"
	"context"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

func NewImportZip(mutex *sync.Mutex, files blob.Store, repo Repository) ImportZip {
	return func(subject auth.Subject, pid ID, src io.Reader) (err error) {
		if err := subject.AuditResource(repo.Name(), string(pid), PermImportZip); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optPrj, err := repo.FindByID(pid)
		if err != nil {
			return err
		}

		if optPrj.IsNone() {
			return os.ErrNotExist
		}

		prj := optPrj.Unwrap()

		// for security, we need to consume and unpack the entire zip file at first, otherwise we
		// may get interrupted in the middle of a restore leaving the system in a broken state, e.g.
		// without a valid user table. This happens often for large backups when uploaded.
		tempFilename := filepath.Join(os.TempDir(), data.RandIdent[string]()+".project.zip")
		tmpFile, err := os.OpenFile(tempFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("could not create temp file: %w", err)
		}
		defer func() {
			_ = tmpFile.Close()
			if err := os.Remove(tempFilename); err != nil {
				slog.Error("cannot remove temp zip file", "err", err)
			}
		}()

		srcSize, err := io.Copy(tmpFile, src)
		if err != nil {
			return fmt.Errorf("could not copy to temp file: %w", err)
		}

		if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("could not seek temp file: %w", err)
		}

		ctx := context.Background()

		// copy is complete
		var dstFiles []File = prj.Files
		zipReader, err := zip.NewReader(tmpFile, srcSize)
		for _, file := range zipReader.File {
			if file.FileHeader.Mode().IsDir() {
				continue
			}

			name := file.Name

			// remove existing file
			dstFiles = slices.DeleteFunc(dstFiles, func(file File) bool {
				if file.Filename == "" {
					// clean up from failed something
					return true
				}
				if file.Filename == name {
					if err := files.Delete(ctx, file.Blob); err != nil {
						slog.Error("cannot delete existing file", "err", err)
					}

					return true
				}

				return false
			})

			reader, err := file.Open()
			if err != nil {
				return fmt.Errorf("could not open zip file: %w", err)
			}

			blobID := data.RandIdent[string]()
			writer, err := files.NewWriter(ctx, blobID)
			if err != nil {
				return err
			}

			if _, err := io.Copy(writer, reader); err != nil {
				_ = writer.Close()
				_ = reader.Close()
				return fmt.Errorf("could not copy to zip file: %w", err)
			}

			_ = reader.Close()

			if err := writer.Close(); err != nil {
				return fmt.Errorf("could not close blob store: %w", err)
			}

			dstFiles = append(dstFiles, File{
				Filename: name,
				Blob:     blobID,
				LastMod:  time.Now(),
			})

			slog.Info("imported project file", "pid", pid, "file", name, "blob", blobID)
		}

		prj.Files = dstFiles

		return repo.Save(prj)
	}
}
