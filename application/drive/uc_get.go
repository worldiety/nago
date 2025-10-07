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

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/presentation/core"
)

func NewGet(repo Repository, blobs blob.Store) Get {
	return func(subject auth.Subject, fid FID, version BID) (option.Opt[core.File], error) {
		optFile, err := readFileStat(repo, fid)
		if err != nil {
			return option.None[core.File](), err
		}

		if optFile.IsNone() {
			return option.None[core.File](), nil
		}

		file := optFile.Unwrap()

		if !file.CanRead(subject) {
			return option.None[core.File](), user.PermissionDeniedErr
		}

		return option.Some[core.File](fileImpl{
			repo:    repo,
			blobs:   blobs,
			fid:     fid,
			version: version,
			file:    file,
		}), nil
	}
}

type fileImpl struct {
	repo    Repository
	blobs   blob.Store
	fid     FID
	version BID
	file    File
}

func (f fileImpl) Name() string {
	return f.file.Filename
}

func (f fileImpl) MimeType() (string, bool) {
	if f.file.FileInfo.IsNone() {
		return "", false
	}

	return f.file.FileInfo.Unwrap().MimeType, true
}

func (f fileImpl) Size() (int64, bool) {
	if f.file.FileInfo.IsNone() {
		return 0, false
	}

	return f.file.FileInfo.Unwrap().Size, true
}

func (f fileImpl) Transfer(dst io.Writer) (int64, error) {
	reader, err := f.Open()
	if err != nil {
		return 0, err
	}

	defer reader.Close()
	return io.Copy(dst, reader)
}

func (f fileImpl) Open() (io.ReadCloser, error) {
	if f.file.IsDir() {
		return nil, fmt.Errorf("is a directory: %w", os.ErrNotExist)
	}

	if f.file.FileInfo.IsNone() {
		return nil, fmt.Errorf("there is no binary data availabe in the file: %w", os.ErrNotExist)
	}

	info := f.file.FileInfo.Unwrap()

	if f.version != "" {
		found := false
		for entry := range f.file.AuditLog.All() {
			if v, ok := entry.Unwrap(); ok {
				if v, ok := v.(VersionAdded); ok && v.FileInfo.Blob == f.version {
					found = true
					break
				}
			}
		}

		if !found {
			return nil, fmt.Errorf("there is no such version: %s: %w", f.version, os.ErrNotExist)
		}
	}

	optReader, err := f.blobs.NewReader(context.Background(), string(info.Blob))
	if err != nil {
		return nil, err
	}

	if optReader.IsNone() {
		return nil, fmt.Errorf("there is no such blob: %s: %w", string(info.Blob), os.ErrNotExist)
	}

	return optReader.Unwrap(), nil
}
