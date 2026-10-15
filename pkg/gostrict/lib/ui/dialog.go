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

// TDialog is a dialog, a window above the content with an optional title, body and footer, see Dialog.
type TDialog struct{ ref js.Ref }

func (TDialog) isView() {}

// Dialog returns a dialog with the body. By default, the dialog has the width L400.
//
//gostrict:host
func Dialog(body View) TDialog {
	return TDialog{ref: wrap(nui.Dialog(viewOf(body)))}
}

// Title sets the title, which is shown at the top.
//
//gostrict:host
func (c TDialog) Title(title View) TDialog {
	return TDialog{ref: wrap(get[nui.TDialog](c.ref).Title(viewOf(title)))}
}

// PreBody sets a view, which is shown before the body.
//
//gostrict:host
func (c TDialog) PreBody(v View) TDialog {
	return TDialog{ref: wrap(get[nui.TDialog](c.ref).PreBody(viewOf(v)))}
}

// TitleX sets an additional title view, e.g. for actions, which is aligned differently from the title.
//
//gostrict:host
func (c TDialog) TitleX(x View) TDialog {
	return TDialog{ref: wrap(get[nui.TDialog](c.ref).TitleX(viewOf(x)))}
}

// Footer sets the footer, e.g. with buttons.
//
//gostrict:host
func (c TDialog) Footer(footer View) TDialog {
	return TDialog{ref: wrap(get[nui.TDialog](c.ref).Footer(viewOf(footer)))}
}

// Alignment aligns the content of the dialog.
//
//gostrict:host
func (c TDialog) Alignment(alignment Alignment) TDialog {
	return TDialog{ref: wrap(get[nui.TDialog](c.ref).Alignment(nui.Alignment(alignment)))}
}

// ModalPadding sets the padding around the dialog.
//
//gostrict:host
func (c TDialog) ModalPadding(padding Padding) TDialog {
	return TDialog{ref: wrap(get[nui.TDialog](c.ref).ModalPadding(padding.nago()))}
}

// Frame sets the layout frame of the dialog, including its size and position.
//
//gostrict:host
func (c TDialog) Frame(frame Frame) TDialog {
	return TDialog{ref: wrap(get[nui.TDialog](c.ref).Frame(frame.nago()))}
}

// WithFrame changes the layout frame of the dialog with fn, which gets the current one.
//
//gostrict:host
func (c TDialog) WithFrame(fn func(Frame) Frame) TDialog {
	return TDialog{ref: wrap(get[nui.TDialog](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }))}
}

// DisableBoxLayout disables the box layout handling of the dialog, if b is true.
//
//gostrict:host
func (c TDialog) DisableBoxLayout(b bool) TDialog {
	return TDialog{ref: wrap(get[nui.TDialog](c.ref).DisableBoxLayout(b))}
}
