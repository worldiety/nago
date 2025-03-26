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

type OptionalFieldsOptions struct {
	Label string

	// Enabled indicates, if the fields shall be shown initially.
	Enabled bool

	// ID overwrites the automatic id creation, which may be wrong if you have the same binding with other
	// Optional fields having the same (or empty) label.
	ID string
}

// OptionalFields creates a kind of invisible section for the given fields.
// A checkbox is shown, to show the given fields.
// This may be used, to simplify complex forms and make them a bit more lightweight if not all fields are required.
func OptionalFields[E any](opts OptionalFieldsOptions, fields ...Field[E]) []Field[E] {
	return fakeFormFields("", optionSection(fields, opts), fields...)
}

func optionSection[E any](fields []Field[E], opts OptionalFieldsOptions) func(bnd *Binding[E], views ...core.View) ui.DecoredView {
	return func(bnd *Binding[E], views ...core.View) ui.DecoredView {
		checkedState := core.StateOf[bool](bnd.wnd, bnd.id+opts.Label+opts.ID).Init(func() bool {
			return opts.Enabled
		})
		cb := ui.Checkbox(checkedState.Get()).InputChecked(checkedState)

		allViews := make([]core.View, 0, len(fields)+1)
		allViews = append(allViews, ui.HStack(
			cb,
			ui.Text(opts.Label),
		))

		if checkedState.Get() {
			allViews = append(allViews, views...)
		}

		return ui.VStack(allViews...).
			Alignment(ui.Leading).
			Frame(ui.Frame{}.FullWidth())
	}
}
