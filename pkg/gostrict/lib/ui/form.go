// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// TForm is a form, which groups input fields and controls, see Form.
type TForm struct{ ref js.Ref }

func (TForm) isView() {}

// Form returns a form, which contains the children.
//
//gostrict:host
func Form(children ...View) TForm {
	return TForm{ref: wrap(nui.Form(viewsOf(children)...))}
}

// ID sets the unique identifier of the form.
//
//gostrict:host
func (c TForm) ID(id string) TForm {
	return TForm{ref: wrap(get[nui.TForm](c.ref).ID(id))}
}

// Action sets the function, which is called, when the form is submitted.
//
//gostrict:host
func (c TForm) Action(action func()) TForm {
	return TForm{ref: wrap(get[nui.TForm](c.ref).Action(action))}
}

// Frame sets the size and the position of the form.
//
//gostrict:host
func (c TForm) Frame(frame Frame) TForm {
	return TForm{ref: wrap(get[nui.TForm](c.ref).Frame(frame.nago()))}
}

// Autocomplete enables or disables the autocomplete of the browser for the form.
//
//gostrict:host
func (c TForm) Autocomplete(b bool) TForm {
	return TForm{ref: wrap(get[nui.TForm](c.ref).Autocomplete(b))}
}
