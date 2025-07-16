// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package image

import (
	"context"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/blob"
	"io"
)

func NewOpenReader(files blob.Store) OpenReader {
	return func(user permission.Auditable, id ID) (option.Opt[io.ReadCloser], error) {
		// TODO solve permission questions

		return files.NewReader(context.Background(), string(id))
	}
}
