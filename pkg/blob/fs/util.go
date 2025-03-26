// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package fs

import (
	"crypto/sha512"
	"io"
	"io/fs"
)

func Sha512_224(fsys fs.FS, name string) ([28]byte, error) {
	r, err := fsys.Open(name)
	if err != nil {
		return [28]byte{}, err
	}

	defer r.Close()

	hasher := sha512.New512_224()
	if _, err := io.Copy(hasher, r); err != nil {
		return [28]byte{}, err
	}

	hash := hasher.Sum(nil)
	var ret [28]byte
	copy(ret[:], hash)
	
	return ret, nil
}
