// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"context"
	"fmt"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xtime"
)

func NewDelete(mutex *sync.Mutex, bus events.Bus, repo Repository, walkDir WalkDir, blobs blob.Store) Delete {
	return func(subject auth.Subject, fid FID, opts DeleteOptions) error {
		mutex.Lock()
		defer mutex.Unlock()

		ctx := context.Background()

		optFile, err := readFileStat(repo, fid)
		if err != nil {
			return fmt.Errorf("cannot read file delete candidate %s: %w", fid, err)
		}

		if optFile.IsNone() {
			// already deleted
			return nil
		}

		file := optFile.Unwrap()
		if file.IsDir() && file.Entries.Len() > 0 && !opts.Recursive {
			return fmt.Errorf("cannot delete file %s as it is a non-empty directory and recursive flag has not been set", fid)
		}

		optParent, err := readFileStat(repo, file.Parent)
		if err != nil {
			return fmt.Errorf("cannot read file delete candidate parent %s: %w", file.Parent, err)
		}

		if !file.CanDelete(subject) {
			return fmt.Errorf("permission denied to delete file %s: %w", fid, user.PermissionDeniedErr)
		}

		if optParent.IsSome() {
			// detach from parent
			parent := optParent.Unwrap()
			parent.Entries = parent.Entries.DeleteFunc(func(f FID) bool {
				return f == fid
			})

			parent.AuditLog = parent.AuditLog.Append(LogEntry{Deleted: option.Pointer(&Deleted{
				FID:    fid,
				ByUser: subject.ID(),
				Time:   xtime.Now(),
			})})

			if err := repo.Save(parent); err != nil {
				return fmt.Errorf("cannot delete file: updating parent failed: %s: %w", parent.ID, err)
			}
		}

		// now find and delete the actual data
		var deleteList []File
		err = walkDir(subject, fid, func(fid FID, file File, err error) error {
			if err != nil {
				return err
			}

			if !file.CanDelete(subject) {
				return fmt.Errorf("permission denied to delete file %s: %w", fid, user.PermissionDeniedErr)
			}

			deleteList = append(deleteList, file)
			return nil
		})

		if err != nil {
			return fmt.Errorf("delete ownership of file tree is incomplete %s: %w", fid, err)
		}

		now := xtime.Now()
		for _, file := range deleteList {
			// purge all blob versions from store
			for _, added := range file.Versions() {
				if err := blobs.Delete(ctx, string(added.FileInfo.Blob)); err != nil {
					return fmt.Errorf("cannot delete blob %s: %w", added.FileInfo.Blob, err)
				}
			}

			if err := repo.DeleteByID(file.ID); err != nil {
				return fmt.Errorf("cannot delete file %s: %w", file.ID, err)
			}

			bus.Publish(Deleted{
				FID:    file.ID,
				ByUser: subject.ID(),
				Time:   now,
			})
		}

		return nil
	}
}
