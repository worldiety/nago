// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewMkDir(mutex *sync.Mutex, repo Repository) MkDir {
	return func(subject auth.Subject, parent FID, name string, opts MkDirOptions) (File, error) {
		var zero File

		if err := ValidateName(name); err != nil {
			return zero, err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optParentFile, err := readFileStat(repo, parent)
		if err != nil {
			return zero, err
		}

		if optParentFile.IsNone() {
			return zero, fmt.Errorf("parent file does not exist: %s: %w", parent, os.ErrNotExist)
		}

		parentFile := optParentFile.Unwrap()

		if !(parentFile.CanWrite(subject) || subject.HasResourcePermission(repo.Name(), string(parent), PermMkDir)) {
			return zero, fmt.Errorf("cannot create directory, either grant write permission to parent or PermMkDir: %w", user.PermissionDeniedErr)
		}

		// try and find if exists
		for fid := range parentFile.Entries.All() {
			optEntryFile, err := readFileStat(repo, fid)
			if err != nil {
				return zero, fmt.Errorf("cannot read child entry file (parent %s) %s: %w", parent, fid, err)
			}

			if optEntryFile.IsNone() {
				slog.Error("stale file reference found", "directory", parent, "entry", fid)
				continue
			}

			entry := optEntryFile.Unwrap()

			if !entry.CanRead(subject) {
				return zero, fmt.Errorf("cannot read child entry file %s: %w", fid, user.PermissionDeniedErr)
			}

			if !entry.IsDir() {
				return zero, fmt.Errorf("child entry exists and is not a directory %s: %w", fid, os.ErrExist)
			}

			if entry.Filename == name {
				return entry, nil
			}
		}

		file := File{
			ID:       data.RandIdent[FID](),
			Filename: name,
			Owner:    opts.User,
			Group:    opts.Group,
			FileMode: os.ModeDir | opts.Mode.Perm(),
			repo:     repo,
		}

		if file.Owner == "" {
			// do not inherit owner, this is the given subject
			file.Owner = subject.ID()
		}

		if file.Group == "" {
			// inherit from context
			file.Group = parentFile.Group
		}

		if opts.Mode == 0 {
			// inherit from context
			file.FileMode = os.ModeDir | parentFile.FileMode
		}

		if optFile, err := repo.FindByID(file.ID); err != nil || optFile.IsSome() {
			if err != nil {
				return zero, err
			}

			return zero, fmt.Errorf("file already exists: %s", file.ID)
		}

		if err := repo.Save(file); err != nil {
			return zero, fmt.Errorf("cannot save file: %w", err)
		}

		parentFile.Entries = parentFile.Entries.Append(file.ID)

		if err := repo.Save(parentFile); err != nil {
			return zero, fmt.Errorf("cannot save parent file: %w", err)
		}

		return file, nil
	}
}
