// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"context"
	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"io"
)

func NewLoadProjectBlob(files blob.Store, repo Repository) LoadProjectBlob {
	return func(subject auth.Subject, pid ID, file BlobID) (option.Opt[io.ReadCloser], error) {
		if err := subject.AuditResource(repo.Name(), string(pid), PermLoadProjectBlob); err != nil {
			return option.None[io.ReadCloser](), err
		}

		optPrj, err := repo.FindByID(pid)
		if err != nil {
			return option.None[io.ReadCloser](), err
		}

		if optPrj.IsNone() {
			return option.None[io.ReadCloser](), nil
		}

		prj := optPrj.Unwrap()
		for _, f := range prj.Files {
			if f.Blob == file {
				return files.NewReader(context.Background(), f.Blob)
			}
		}

		return option.None[io.ReadCloser](), nil
	}
}
