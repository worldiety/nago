// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/presentation/core"
)

func NewZip(repo Repository, blobs blob.Store, walkDir WalkDir) Zip {
	return func(subject auth.Subject, fids []FID) (core.File, error) {
		return zipFile{
			repo:    repo,
			blobs:   blobs,
			fids:    fids,
			subject: subject,
			walkDir: walkDir,
		}, nil
	}
}

type zipFile struct {
	repo    Repository
	blobs   blob.Store
	fids    []FID
	subject auth.Subject
	walkDir WalkDir
}

func (z zipFile) Name() string {
	return "files.zip"
}

func (z zipFile) MimeType() (string, bool) {
	return "application/zip", true
}

func (z zipFile) Size() (int64, bool) {
	return 0, false
}

func (z zipFile) Transfer(dst io.Writer) (int64, error) {
	// collect all recursive candidates
	var files []File
	for _, fid := range z.fids {
		err := z.walkDir(z.subject, fid, func(fid FID, file File, err error) error {
			if err != nil {
				return err
			}

			if !file.CanRead(z.subject) {
				return fmt.Errorf("file %s is not readable: %w", fid, user.PermissionDeniedErr)
			}

			if file.IsDir() {
				return nil
			}

			files = append(files, file)
			return nil
		})

		if err != nil {
			return 0, err
		}
	}

	cw := &countingWriter{W: dst}
	zipWriter := zip.NewWriter(cw)
	for _, file := range files {

		path, err := file.AbsolutePath()
		if err != nil {
			return cw.Count, err
		}

		f, err := zipWriter.Create(path)
		if err != nil {
			return cw.Count, err
		}

		if file.FileInfo.IsNone() {
			// zero byte file
			continue
		}

		key := string(file.FileInfo.Unwrap().Blob)
		optReader, err := z.blobs.NewReader(context.Background(), key)
		if err != nil {
			return cw.Count, err
		}

		if optReader.IsNone() {
			return cw.Count, fmt.Errorf("blod not found: %s: %w", key, os.ErrNotExist)
		}

		reader := optReader.Unwrap()
		if _, err := io.Copy(f, reader); err != nil {
			return cw.Count, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return cw.Count, err
	}

	return cw.Count, nil
}

func (z zipFile) Open() (io.ReadCloser, error) {
	return nil, fmt.Errorf("this implementation does not support open, use Transfer")
}

type countingWriter struct {
	W     io.Writer
	Count int64
}

func (cw *countingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.W.Write(p)
	cw.Count += int64(n)
	return
}
