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

// Label adds the given text as a static text into the form
func Label[T any](label string) Field[T] {
	return Field[T]{
		RenderFormElement: func(self Field[T], entity *core.State[T]) ui.DecoredView {
			return ui.VStack(ui.Text(label)).Alignment(ui.Leading).
				Frame(ui.Frame{}.FullWidth())

		},
	}
}
