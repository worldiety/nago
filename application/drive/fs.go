// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import "io/fs"

type fsImpl struct {
}

func (f fsImpl) Stat(name string) (fs.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (f fsImpl) Open(name string) (fs.File, error) {
	//TODO implement me
	panic("implement me")
}

type fileReadSeekCloser struct {
}

func (f fileReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (f fileReadSeekCloser) Stat() (fs.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (f fileReadSeekCloser) Read(bytes []byte) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (f fileReadSeekCloser) Close() error {
	//TODO implement me
	panic("implement me")
}
