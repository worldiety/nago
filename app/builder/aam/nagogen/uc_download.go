// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagogen

import (
	"archive/zip"
	"bytes"

	"go.wdy.de/nago/app/builder/aam"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

func NewDownload(gen Generate) Download {
	return func(subject auth.Subject, model *aam.App) (core.File, error) {
		fsys, err := gen(subject, model)
		if err != nil {
			return nil, err
		}
		var tmp bytes.Buffer
		w := zip.NewWriter(&tmp)
		if err := w.AddFS(fsys); err != nil {
			return nil, err
		}

		if err := w.Close(); err != nil {
			return nil, err
		}

		return core.MemFile{
			Filename:     string(model.ID) + ".zip",
			MimeTypeHint: "application/zip",
			Bytes:        tmp.Bytes(),
		}, nil
	}
}
