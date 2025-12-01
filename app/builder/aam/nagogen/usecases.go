// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagogen

import (
	"io/fs"

	"go.wdy.de/nago/app/builder/aam"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

type Generate func(subject auth.Subject, model *aam.App) (fs.FS, error)
type Download func(subject auth.Subject, model *aam.App) (core.File, error)

type UseCases struct {
	Generate Generate
	Download Download
}

func NewUseCases() UseCases {
	genFn := NewGenerate()
	return UseCases{
		Generate: genFn,
		Download: NewDownload(genFn),
	}
}
