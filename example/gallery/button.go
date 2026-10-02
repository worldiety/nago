// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	. "go.wdy.de/nago/presentation/ui"
)

func init() {
	register("button", func(wnd core.Window) core.View {
		return HStack(
			PrimaryButton(func() {}).Title("Primary"),
			SecondaryButton(func() {}).Title("Secondary"),
			TertiaryButton(func() {}).Title("Tertiary"),
			PrimaryButton(func() {}).Title("With icon").PreIcon(icons.SpeakerWave),
		).Gap(L16)
	})
}
