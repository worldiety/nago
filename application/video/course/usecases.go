// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package course

import (
	"go.wdy.de/nago/application/drive"
	"go.wdy.de/nago/application/video"
)

type ID string

type Course struct {
	ID    ID
	Title string
}

type Chapter struct {
	Entries []drive.FID
}

type Entry struct {
	Video video.ID
	PDF   drive.FID
}
