// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidataimport

import (
	"go.wdy.de/nago/application/dataimport"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func PageTransformation(wnd core.Window, ucImp dataimport.UseCases) core.View {
	return ui.VStack(
		ui.H1("Transformation"),
	).Alignment(ui.Leading).
		FullWidth()
}
