// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"strings"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func RenderPlaintext(ctx FieldContext) core.View {
	if strings.HasPrefix(ctx.Field().Name, "_") && ctx.Label() != "_" {
		return ui.Text(ctx.Label()).FullWidth().TextAlignment(ui.TextAlignStart)
	}
	
	return nil
}
