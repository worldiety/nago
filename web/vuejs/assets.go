// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vuejs

import (
	"embed"
	"io/fs"
)

//go:embed  dist/index.html dist/legacy/* dist/modern/*
var Frontend embed.FS

func Dist() fs.FS {
	fsys, err := fs.Sub(Frontend, "dist")
	if err != nil {
		panic(err)
	}
	return fsys
}
