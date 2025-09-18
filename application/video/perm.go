// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package video

import "go.wdy.de/nago/application/permission"

var (
	PermCreate = permission.Declare[Create]("nago.video.create", "Video erstellen", "Träger dieser Berechtigung können neue Videos erstellen")
)
