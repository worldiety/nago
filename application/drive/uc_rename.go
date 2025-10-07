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
	"go.wdy.de/nago/pkg/xtime"
)

func NewRename(mutex *sync.Mutex, bus events.Bus, repo Repository) Rename {
	return func(subject auth.Subject, fid FID, newName string) error {
		mutex.Lock()
		defer mutex.Unlock()

		if err := ValidateName(newName); err != nil {
			return err
		}

		optFile, err := readFileStat(repo, fid)
		if err != nil {
			return err
		}

		if optFile.IsNone() {
			return os.ErrNotExist
		}

		file := optFile.Unwrap()

		if file.Filename == newName {
			// no-op
			return nil
		}

		optParent, err := readFileStat(repo, file.Parent)
		if err != nil {
			return fmt.Errorf("cannot read parent file: %s: %w", file.Parent, err)
		}

		if optParent.IsNone() {
			return fmt.Errorf("cannot rename a file without a parent: %w", os.ErrNotExist)
		}

		parent := optParent.Unwrap()

		if !file.CanRename(subject) && !parent.CanRename(subject) {
			return user.PermissionDeniedErr
		}

		optEntry, err := parent.EntryByName(newName)
		if err != nil {
			return err
		}

		if optEntry.IsSome() {
			return fmt.Errorf("cannot rename file, because another file already exists in parent directory: %q -> %q: %w", file.Name(), newName, os.ErrExist)
		}

		log := Renamed{
			Name:    newName,
			ByUser:  subject.ID(),
			ModTime: xtime.Now(),
		}

		file.Filename = newName
		file.AuditLog = file.AuditLog.Append(LogEntry{Renamed: option.Pointer(&log)})

		if err := repo.Save(file); err != nil {
			return err
		}

		bus.Publish(log)

		return nil
	}
}
