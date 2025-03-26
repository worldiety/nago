// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// HLine adds a horizontal separator into the form.
func HLine[T any]() Field[T] {
	return Field[T]{
		RenderFormElement: func(self Field[T], entity *core.State[T]) ui.DecoredView {
			return ui.VStack(ui.HLine().Padding(ui.Padding{})).
				Frame(ui.Frame{}.FullWidth())

		},
	}
}
