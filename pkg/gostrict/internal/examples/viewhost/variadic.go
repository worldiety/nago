// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package viewhost

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

// Variadic passes views of several types to the host in the list of a variadic parameter of the type ui.View.
func Variadic(dp float64) string {
	w := &widths{}
	_ = ui.VStack(
		ui.Text("a").Frame(frame(dp)).WithFrame(w.record()),
		ui.ImageIcon("<svg/>"),
		ui.HLine(),
		ui.Spacer(),
		ui.AvatarText("AB"),
		ui.QrCode("q"),
	).Frame(frame(2 * dp)).WithFrame(w.record())
	return w.s
}
