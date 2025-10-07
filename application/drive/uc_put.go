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
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
)

func NewPut(mutex *sync.Mutex, bus events.Bus, repo Repository, blobs blob.Store) Put {
	return func(subject auth.Subject, parent FID, name string, src io.Reader, opts PutOptions) error {
		mutex.Lock()
		defer mutex.Unlock()

		optParentFile, err := readFileStat(repo, parent)
		if err != nil {
			return fmt.Errorf("cannot open parent file: %w", err)
		}

		if optParentFile.IsNone() {
			return fmt.Errorf("parent file does not exist: %s: %w", parent, os.ErrNotExist)
		}

		parentFile := optParentFile.Unwrap()
		if !parentFile.IsDir() {
			return fmt.Errorf("parent file is not a directory: %s: %w", parent, os.ErrInvalid)
		}

		if !parentFile.CanWrite(subject) {
			return fmt.Errorf("cannot write to parent file: %s: %w", parent, user.PermissionDeniedErr)
		}

		optFile, err := parentFile.EntryByName(name)
		if err != nil {
			return fmt.Errorf("cannot get entry by name: %s: %w", name, err)
		}

		if opts.Owner == "" {
			opts.Owner = parentFile.Owner
		}

		if opts.Group == "" {
			opts.Group = parentFile.Group
		}

		if opts.Mode == 0 {
			opts.Mode = parentFile.FileMode.Perm()
		}

		requiresParentFlush := false
		if optFile.IsNone() {
			mode := opts.Mode.Perm()
			file := File{
				ID:       data.RandIdent[FID](),
				FileMode: mode,
				Group:    opts.Group,
				repo:     repo,
				Owner:    opts.Owner,
				Parent:   parent,
				AuditLog: xslices.Wrap[LogEntry](LogEntry{Created: option.Pointer(&Created{
					Owner:    opts.Owner,
					Group:    opts.Group,
					FileMode: mode,
					ByUser:   subject.ID(),
					Time:     xtime.Now(),
				})}),
			}

			if optFile, err := repo.FindByID(file.ID); optFile.IsSome() || err != nil {
				if err != nil {
					return fmt.Errorf("cannot check file: %s: %w", file.ID, err)
				}

				return fmt.Errorf("file already exists: %s: %w", file.ID, os.ErrExist)
			}

			requiresParentFlush = true

			optFile = option.Some(file)
		}

		key, size, err := storeBlob(blobs, src)
		if err != nil {
			return fmt.Errorf("cannot store blob: %w", err)
		}

		shaHash, err := hash(blobs, key)

		mime, err := mimeType(blobs, key)
		if err != nil {
			return fmt.Errorf("cannot detect mime type: %w", err)
		}

		now := xtime.Now()
		file := optFile.Unwrap()
		versionAdded := VersionAdded{
			SourceHint: opts.SourceHint,
			FileInfo: FileInfo{
				OriginalFilename: opts.OriginalFilename,
				Blob:             key,
				Sha3H256:         shaHash,
				Size:             size,
				MimeType:         mime,
			},
			ByUser: subject.ID(),
			Time:   now,
		}

		if file.Filename == "" {
			file.Filename = versionAdded.FileInfo.OriginalFilename
		}

		if file.Filename == "" {
			file.Filename = versionAdded.FileInfo.Blob
		}

		file.FileInfo = option.Some(versionAdded.FileInfo)
		file.AuditLog = file.AuditLog.Append(LogEntry{VersionAdded: option.Pointer(&versionAdded)})

		if err := repo.Save(file); err != nil {
			return fmt.Errorf("cannot save file: %w", err)
		}

		if v, ok := file.AuditLog.Last(); ok {
			if v, ok := v.Unwrap(); ok {
				bus.Publish(v)
			}
		}

		if requiresParentFlush {
			parentFile.Entries = parentFile.Entries.Append(file.ID)
			parentFile.AuditLog = parentFile.AuditLog.Append(LogEntry{Added: option.Pointer(&Added{
				FID:    file.ID,
				ByUser: subject.ID(),
				Time:   now,
			})})

			sortedEntries, err := applyStandardEntryOrder(repo, parentFile.Entries.All())
			if err != nil {
				return err
			}
			parentFile.Entries = xslices.Wrap(sortedEntries...)

			if err := repo.Save(parentFile); err != nil {
				return fmt.Errorf("cannot save parent file: %s: %w", parentFile.ID, err)
			}

			if v, ok := parentFile.AuditLog.Last(); ok {
				if v, ok := v.Unwrap(); ok {
					bus.Publish(v)
				}
			}

		}

		return nil
	}
}

func storeBlob(store blob.Store, src io.Reader) (string, int64, error) {
	blobKey := data.RandIdent[string]()
	if ok, err := store.Exists(context.Background(), blobKey); ok || err != nil {
		if err != nil {
			return blobKey, 0, err
		}

		return blobKey, 0, fmt.Errorf("blob already exists: %s: %w", blobKey, err)
	}

	n, err := blob.Write(store, blobKey, src)
	if err != nil {
		return blobKey, 0, err
	}

	return blobKey, n, nil
}

func hash(store blob.Store, key string) (Sha3H256, error) {
	optReader, err := store.NewReader(context.Background(), key)
	if err != nil {
		return "", err
	}

	if optReader.IsNone() {
		return "", os.ErrNotExist
	}

	reader := optReader.Unwrap()
	defer reader.Close()

	return NewSha3H256(reader)
}

func mimeType(store blob.Store, key string) (string, error) {
	type filer interface {
		LocalFile(key string) (option.Opt[string], error)
	}

	if lf, ok := store.(filer); ok {
		optFname, err := lf.LocalFile(key)
		if err != nil {
			return "", err
		}

		if optFname.IsNone() {
			return "", os.ErrNotExist
		}

		fname := optFname.Unwrap()

		cmd := exec.Command("file", "-I", fname)
		buf, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("cannot exec file cmd %s: %s, %w", fname, string(buf), err)
		}

		tokens := strings.SplitN(string(buf), ":", 2)
		if len(tokens) != 2 {
			return "", fmt.Errorf("cannot process exec file cmd result %s: %s, %w", fname, string(buf), err)
		}

		return strings.TrimSpace(tokens[1]), nil
	}

	return "", fmt.Errorf("cannot get system file tool to detect mimetype %s: %w", key, os.ErrNotExist)
}
