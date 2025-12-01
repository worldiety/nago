// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagogen

import (
	"fmt"
	"io/fs"
	"testing/fstest"

	"go.wdy.de/nago/app/builder/aam"
	"go.wdy.de/nago/auth"
)

func NewGenerate() Generate {
	return func(subject auth.Subject, model *aam.App) (fs.FS, error) {
		fsys := fstest.MapFS{}
		if !model.ID.Valid() {
			return fsys, fmt.Errorf("invalid app id %s", model.ID)
		}

		if err := genMain(fsys, model); err != nil {
			return nil, err
		}

		if err := genMod(fsys, model); err != nil {
			return nil, err
		}

		return fsys, nil
	}
}
