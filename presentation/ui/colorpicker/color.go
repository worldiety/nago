// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package colorpicker

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// TColor is a utility component(Color).
// It wraps a ui.Color and can be rendered to show its visual representation.
type TColor struct {
	color ui.Color
}

// Color creates a new TColor component with the given color.
func Color(color ui.Color) TColor {
	return TColor{color: color}
}

// Render displays the TColor by delegating to renderColor.
func (c TColor) Render(ctx core.RenderContext) core.RenderNode {
	return renderColor(nil, c.color).Render(ctx)
}
