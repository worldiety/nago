// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package navsplitview

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// paneCell returns the grid cell of a column which shows a view of the split view.
//
// A grid item has an automatic minimum width of its content, so a flexible track like 1fr never becomes
// narrower than the widest thing inside, e.g. a table or a row which does not wrap. Such a column grows beyond
// the viewport, where the grid clips it, and nested scroll views never scroll because they are never narrower
// than their content. Therefore, the view is wrapped into a stack which fills the cell and may shrink below its
// content. The stack places the view like the cell did before, so relative heights still refer to the cell.
func paneCell(view core.View, alignment ui.Alignment) ui.TGridCell {
	return ui.GridCell(
		ui.VStack(view).
			Alignment(alignment).
			Frame(ui.Frame{Width: ui.Full, Height: ui.Full, MinWidth: "0px"}),
	).Alignment(ui.Stretch)
}
