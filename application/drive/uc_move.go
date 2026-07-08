// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"fmt"
	"os"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
)

func NewMove(mutex *sync.Mutex, bus events.Bus, repo Repository, walkDir WalkDir) Move {
	return func(subject auth.Subject, fid FID, newParent FID) error {
		mutex.Lock()
		defer mutex.Unlock()

		if fid == "" {
			return fmt.Errorf("cannot move empty file id: %w", os.ErrInvalid)
		}

		if fid == newParent {
			return fmt.Errorf("cannot move a file into itself: %s: %w", fid, os.ErrInvalid)
		}

		optFile, err := readFileStat(repo, fid)
		if err != nil {
			return fmt.Errorf("cannot read file to move: %s: %w", fid, err)
		}

		if optFile.IsNone() {
			return fmt.Errorf("file to move does not exist: %s: %w", fid, os.ErrNotExist)
		}

		file := optFile.Unwrap()

		if file.Parent == "" {
			return fmt.Errorf("cannot move a file without a parent (e.g. a drive root): %s: %w", fid, os.ErrInvalid)
		}

		if file.Parent == newParent {
			// no-op: already located in the requested parent
			return nil
		}

		optOldParent, err := readFileStat(repo, file.Parent)
		if err != nil {
			return fmt.Errorf("cannot read old parent: %s: %w", file.Parent, err)
		}

		if optOldParent.IsNone() {
			return fmt.Errorf("old parent does not exist: %s: %w", file.Parent, os.ErrNotExist)
		}

		optNewParent, err := readFileStat(repo, newParent)
		if err != nil {
			return fmt.Errorf("cannot read destination parent: %s: %w", newParent, err)
		}

		if optNewParent.IsNone() {
			return fmt.Errorf("destination parent does not exist: %s: %w", newParent, os.ErrNotExist)
		}

		oldParent := optOldParent.Unwrap()
		newParentFile := optNewParent.Unwrap()

		if !newParentFile.IsDir() {
			return fmt.Errorf("destination is not a directory: %s: %w", newParent, os.ErrInvalid)
		}

		// unix rename semantics: write on the old parent (to remove the entry) and write on the destination
		// (to add the entry). The moved file itself does not need to be writable.
		if !oldParent.CanWrite(subject) || !newParentFile.CanWrite(subject) {
			return fmt.Errorf("moving requires write permission on both the source and the destination directory: %w", user.PermissionDeniedErr)
		}

		// cycle protection: moving a directory into itself or one of its own descendants would detach the
		// subtree. We traverse the source subtree with system privileges (this is a structural integrity
		// check, not an access decision) and reject if the destination is contained within it.
		if file.IsDir() {
			cycle := false
			err = walkDir(user.SU(), fid, func(walkedFID FID, _ File, werr error) error {
				if werr != nil {
					return werr
				}
				if walkedFID == newParent {
					cycle = true
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("cannot verify move target: %s: %w", newParent, err)
			}
			if cycle {
				return fmt.Errorf("cannot move a directory into itself or one of its descendants: %s -> %s: %w", fid, newParent, os.ErrInvalid)
			}
		}

		// name collision in the destination directory
		optEntry, err := newParentFile.EntryByName(file.Filename)
		if err != nil {
			return fmt.Errorf("cannot check destination for name collision: %w", err)
		}

		if optEntry.IsSome() {
			return fmt.Errorf("a file with the same name already exists in the destination: %q: %w", file.Filename, os.ErrExist)
		}

		now := xtime.Now()

		// 1. detach from the old parent
		oldParent.Entries = oldParent.Entries.DeleteFunc(func(f FID) bool {
			return f == fid
		})
		oldParent.AuditLog = oldParent.AuditLog.Append(LogEntry{Deleted: option.Pointer(&Deleted{
			FID:    fid,
			ByUser: subject.ID(),
			Time:   now,
		})})

		sortedOld, err := applyStandardEntryOrder(repo, oldParent.Entries.All())
		if err != nil {
			return err
		}
		oldParent.Entries = xslices.Wrap(sortedOld...)

		if err := repo.Save(oldParent); err != nil {
			return fmt.Errorf("cannot save old parent: %s: %w", oldParent.ID, err)
		}

		// 2. reparent the file itself
		moved := Moved{
			FID:       fid,
			OldParent: file.Parent,
			NewParent: newParent,
			ByUser:    subject.ID(),
			Time:      now,
		}
		file.Parent = newParent
		file.AuditLog = file.AuditLog.Append(LogEntry{Moved: option.Pointer(&moved)})

		if err := repo.Save(file); err != nil {
			return fmt.Errorf("cannot save moved file: %s: %w", file.ID, err)
		}

		// 3. attach to the destination parent
		newParentFile.Entries = newParentFile.Entries.Append(fid)
		newParentFile.AuditLog = newParentFile.AuditLog.Append(LogEntry{Added: option.Pointer(&Added{
			FID:    fid,
			ByUser: subject.ID(),
			Time:   now,
		})})

		sortedNew, err := applyStandardEntryOrder(repo, newParentFile.Entries.All())
		if err != nil {
			return err
		}
		newParentFile.Entries = xslices.Wrap(sortedNew...)

		if err := repo.Save(newParentFile); err != nil {
			return fmt.Errorf("cannot save destination parent: %s: %w", newParentFile.ID, err)
		}

		bus.Publish(moved)
		if log, ok := oldParent.AuditLog.Last(); ok {
			if v, ok := log.Unwrap(); ok {
				bus.Publish(v)
			}
		}
		if log, ok := newParentFile.AuditLog.Last(); ok {
			if v, ok := log.Unwrap(); ok {
				bus.Publish(v)
			}
		}

		return nil
	}
}
