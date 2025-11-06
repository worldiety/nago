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

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
)

func NewOpenDrive(mutex *sync.Mutex, repo Repository, globalRootRepo NamedRootRepository, userRootRepo UserRootRepository) OpenDrive {
	return func(subject auth.Subject, opts OpenDriveOptions) (Drive, error) {
		var root FID
		var zero Drive

		if opts.Name == "" {
			opts.Name = FSDrive
		}

		mutex.Lock()
		defer mutex.Unlock()

		switch opts.Namespace {
		case NamespacePrivate:
			id, err := openUserRoot(repo, userRootRepo, subject, opts)
			if err != nil {
				return zero, fmt.Errorf("failed top open user root: %w", err)
			}

			root = id
		case NamespaceGlobal:
			id, err := openGlobalRoot(repo, globalRootRepo, subject, opts)
			if err != nil {
				return zero, fmt.Errorf("failed top open global root: %w", err)
			}

			root = id
		default:
			return zero, fmt.Errorf("unknown origin: %v", opts.Namespace)
		}

		optFile, err := repo.FindByID(root)
		if err != nil {
			return zero, fmt.Errorf("cannot open root: %s: %w", root, err)
		}

		if optFile.IsNone() {
			return zero, fmt.Errorf("referenced root file does not exist: %s: %w", root, err)
		}

		file := optFile.Unwrap()

		if opts.Create && user.IsSU(subject) {
			chmodRequired := file.Group != opts.Group || file.FileMode.Perm() != opts.Mode.Perm()
			if chmodRequired {
				slog.Warn("su changed root permissions", "fid", file.ID)
				file.Group = opts.Group
				file.FileMode = opts.Mode
				if err := repo.Save(file); err != nil {
					return zero, fmt.Errorf("failed su chmod file: %w", err)
				}
			}
		}

		if !file.CanRead(subject) {
			return zero, fmt.Errorf("user is not allowed to read file: %s: %w", subject.ID(), user.PermissionDeniedErr)
		}

		return Drive{
			Name:      opts.Name,
			Root:      file.ID,
			Namespace: opts.Namespace,
		}, nil
	}
}

func openGlobalRoot(repo Repository, globalRootRepo NamedRootRepository, subject auth.Subject, opts OpenDriveOptions) (FID, error) {
	if err := subject.AuditResource(globalRootRepo.Name(), opts.Name, PermOpenFile); err != nil {
		return "", err
	}

	optGlobalRoot, err := globalRootRepo.FindByID(opts.Name)
	if err != nil {
		return "", fmt.Errorf("failed find global root: %s: %w", opts.Name, err)
	}

	if optGlobalRoot.IsNone() {
		if !opts.Create {
			return "", fmt.Errorf("the global drive root doesn't exist: %s: %w", opts.Name, os.ErrNotExist)
		}

		optGlobalRoot = option.Some(NamedRoot{
			ID: opts.Name,
		})
	}

	globalRoot := optGlobalRoot.Unwrap()
	if globalRoot.Root == "" {
		tmp := newRandFileFromOpts(repo, subject, opts)
		if optFile, err := repo.FindByID(tmp.ID); err != nil || optFile.IsSome() {
			if err != nil {
				return "", fmt.Errorf("failed to create new global drive root: %s: %w", opts.Name, os.ErrExist)
			}

			if optFile.IsSome() {
				return "", fmt.Errorf("the global drive root already exists: %s: %w", tmp.ID, os.ErrExist)
			}
		}

		if err := repo.Save(tmp); err != nil {
			return "", fmt.Errorf("failed to save new global drive root: %s: %s: %w", opts.Name, tmp.ID, err)
		}

		globalRoot.Root = tmp.ID

		if err := globalRootRepo.Save(globalRoot); err != nil {
			return "", fmt.Errorf("failed to save global drive root: %s: %w", opts.Name, err)
		}
	}

	return globalRoot.Root, nil
}

func openUserRoot(repo Repository, userRootRepo UserRootRepository, subject auth.Subject, opts OpenDriveOptions) (FID, error) {
	if err := subject.AuditResource(userRootRepo.Name(), string(opts.User), PermOpenFile); err != nil {
		return "", err
	}

	optUsrRoots, err := userRootRepo.FindByID(opts.User)
	if err != nil {
		return "", fmt.Errorf("")
	}

	if optUsrRoots.IsNone() {
		if !opts.Create {
			return "", fmt.Errorf("the user drive root doesn't exist: %s: %w", opts.User, os.ErrNotExist)
		}

		tmp := UserRoots{
			ID:    opts.User,
			Roots: map[string]FID{},
		}

		optUsrRoots = option.Some(tmp)
	}

	usrRoot := optUsrRoots.Unwrap()
	rootID, ok := usrRoot.Roots[opts.Name]
	if !ok && !opts.Create {
		return "", fmt.Errorf("the user drive root doesn't contain the named root: %s: %w", opts.Name, os.ErrNotExist)
	}

	if rootID == "" {
		tmp := newRandFileFromOpts(repo, subject, opts)

		if optFile, err := repo.FindByID(tmp.ID); err != nil || optFile.IsSome() {
			if err != nil {
				return "", fmt.Errorf("failed to create new user drive root: %s: %w", opts.Name, os.ErrExist)
			}

			if optFile.IsSome() {
				return "", fmt.Errorf("the user drive root already exists: %s: %w", tmp.ID, os.ErrExist)
			}
		}

		if err := repo.Save(tmp); err != nil {
			return "", fmt.Errorf("failed to save new user drive root: %s: %s: %w", opts.Name, tmp.ID, err)
		}

		if usrRoot.Roots == nil {
			usrRoot.Roots = map[string]FID{}
		}

		usrRoot.Roots[opts.Name] = tmp.ID
		rootID = tmp.ID
	}

	if err := userRootRepo.Save(usrRoot); err != nil {
		return "", fmt.Errorf("failed to save user drive roots: %s: %w", opts.User, err)
	}

	return rootID, nil
}

func newRandFileFromOpts(repo Repository, subject auth.Subject, opts OpenDriveOptions) File {
	mode := os.ModeDir | opts.Mode.Perm()
	return File{
		ID:       data.RandIdent[FID](),
		FileMode: mode,
		Group:    opts.Group,
		repo:     repo,
		Owner:    opts.User,
		AuditLog: xslices.Wrap[LogEntry](LogEntry{Created: option.Pointer(&Created{
			Owner:    opts.User,
			Group:    opts.Group,
			FileMode: mode,
			ByUser:   subject.ID(),
			Time:     xtime.Now(),
		})}),
	}
}
