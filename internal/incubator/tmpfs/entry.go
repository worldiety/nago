// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tmpfs

import (
	"encoding/json"
	"io/fs"
	"os"
)

// DirEntry represents a temporary file in the applications scope.
// Its origin is usually from a different system, process or server.
// This may be some file upload through a webbrowser or a resolved file from an Android Content Provider or an
// image from an iOS PhotoAsset file.
type DirEntry struct {
	name             string
	absoluteMetaPath string
}

func NewDirEntry(name string, absoluteMetaPath string) *DirEntry {
	return &DirEntry{name: name, absoluteMetaPath: absoluteMetaPath}
}

// Name returns the local temporary name of the file. This avoids a violation against the fs.FS abstraction.
// Note, that when representing URLs or S3 paths it does not make sense to use the fs semantics here, as their
// segments do not represent directories either. Thus, we do not model that here.
// To get the original resource name, if any, use [tmpfs.FileInfo.ResourceName].
func (d DirEntry) Name() string {
	return d.name
}

func (d DirEntry) IsDir() bool {
	return false
}

func (d DirEntry) Type() fs.FileMode {
	return 0 // always a regular file
}

func (d DirEntry) Info() (fs.FileInfo, error) {
	buf, err := os.ReadFile(d.absoluteMetaPath)
	if err != nil {
		return nil, err
	}

	var info FileInfo
	if err := json.Unmarshal(buf, &info); err != nil {
		return nil, err
	}

	return info, nil
}
