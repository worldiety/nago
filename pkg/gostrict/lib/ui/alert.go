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
	"go.wdy.de/nago/presentation/ui/alert"
)

// AlertOption is an option of AlertDialog, e.g. a button or a width. Like alert.Option of nago.
type AlertOption struct{ ref js.Ref }

// AlertOk adds a default button, which closes the dialog.
//
//gostrict:host
func AlertOk() AlertOption {
	return AlertOption{ref: wrap(alert.Ok())}
}

// AlertDelete adds a button, which closes the dialog and calls onDelete.
//
//gostrict:host
func AlertDelete(onDelete func()) AlertOption {
	return AlertOption{ref: wrap(alert.Delete(onDelete))}
}

// AlertSave adds a button, which calls onSave and closes the dialog, if onSave returns true.
//
//gostrict:host
func AlertSave(onSave func() bool) AlertOption {
	return AlertOption{ref: wrap(alert.Save(onSave))}
}

// AlertCreate adds a button, which calls onSave and closes the dialog, if onSave returns true.
//
//gostrict:host
func AlertCreate(onSave func() bool) AlertOption {
	return AlertOption{ref: wrap(alert.Create(onSave))}
}

// AlertAdd adds a button, which calls onSave and closes the dialog, if onSave returns true.
//
//gostrict:host
func AlertAdd(onSave func() bool) AlertOption {
	return AlertOption{ref: wrap(alert.Add(onSave))}
}

// AlertApply adds a button, which calls onSave and closes the dialog, if onSave returns true.
//
//gostrict:host
func AlertApply(onSave func() bool) AlertOption {
	return AlertOption{ref: wrap(alert.Apply(onSave))}
}

// AlertClose adds a button, which calls onSave and closes the dialog, if onSave returns true.
//
//gostrict:host
func AlertClose(onSave func() bool) AlertOption {
	return AlertOption{ref: wrap(alert.Close(onSave))}
}

// AlertConfirm adds a button, which calls onSave and closes the dialog, if onSave returns true.
//
//gostrict:host
func AlertConfirm(onSave func() bool) AlertOption {
	return AlertOption{ref: wrap(alert.Confirm(onSave))}
}

// AlertCloseable adds a close icon button, which closes the dialog, and lets the escape key close it.
//
//gostrict:host
func AlertCloseable() AlertOption {
	return AlertOption{ref: wrap(alert.Closeable())}
}

// AlertMinWidth sets the minimum width of the dialog. It is probably not what you want, use AlertWidth instead.
//
//gostrict:host
func AlertMinWidth(w Length) AlertOption {
	return AlertOption{ref: wrap(alert.MinWidth(nui.Length(w)))}
}

// AlertWidth sets the width of the dialog.
//
//gostrict:host
func AlertWidth(w Length) AlertOption {
	return AlertOption{ref: wrap(alert.Width(nui.Length(w)))}
}

// AlertLarge sets the width of the dialog to 560dp (35rem).
//
//gostrict:host
func AlertLarge() AlertOption {
	return AlertOption{ref: wrap(alert.Large())}
}

// AlertLarger sets the width of the dialog to 880dp (55rem).
//
//gostrict:host
func AlertLarger() AlertOption {
	return AlertOption{ref: wrap(alert.Larger())}
}

// AlertXLarge sets the width of the dialog to 1200dp (75rem).
//
//gostrict:host
func AlertXLarge() AlertOption {
	return AlertOption{ref: wrap(alert.XLarge())}
}

// AlertXXLarge sets the width of the dialog to 1600dp (100rem).
//
//gostrict:host
func AlertXXLarge() AlertOption {
	return AlertOption{ref: wrap(alert.XXLarge())}
}

// AlertFullHeight lets the dialog take the full available height.
//
//gostrict:host
func AlertFullHeight() AlertOption {
	return AlertOption{ref: wrap(alert.FullHeight())}
}

// AlertHeight sets the height of the dialog.
//
//gostrict:host
func AlertHeight(h Length) AlertOption {
	return AlertOption{ref: wrap(alert.Height(nui.Length(h)))}
}

// AlertCancel adds a button, which closes the dialog and calls onCancel.
//
//gostrict:host
func AlertCancel(onCancel func()) AlertOption {
	return AlertOption{ref: wrap(alert.Cancel(onCancel))}
}

// AlertBack adds a button, which closes the dialog and calls onCancel.
//
//gostrict:host
func AlertBack(onCancel func()) AlertOption {
	return AlertOption{ref: wrap(alert.Back(onCancel))}
}

// AlertAlignment sets the alignment of the dialog content.
//
//gostrict:host
func AlertAlignment(alignment Alignment) AlertOption {
	return AlertOption{ref: wrap(alert.Alignment(nui.Alignment(alignment)))}
}

// AlertModalPadding sets the padding inside the modal dialog.
//
//gostrict:host
func AlertModalPadding(padding Padding) AlertOption {
	return AlertOption{ref: wrap(alert.ModalPadding(padding.nago()))}
}

// AlertPreBody sets a view between the title and the body, which does not scroll.
//
//gostrict:host
func AlertPreBody(v View) AlertOption {
	return AlertOption{ref: wrap(alert.PreBody(viewOf(v)))}
}

// TAlertDialog is a modal dialog, which is shown above the window, see AlertDialog.
type TAlertDialog struct{ ref js.Ref }

func (TAlertDialog) isView() {}

// AlertDialog returns a modal dialog with the title, the body and the options, which is shown while isPresented is
// true. Like alert.Dialog of nago.
//
//gostrict:host
func AlertDialog(title string, body View, isPresented BoolState, opts ...AlertOption) TAlertDialog {
	return TAlertDialog{ref: wrap(alert.Dialog(title, viewOf(body), isPresented.nago(), alertOptionsOf(opts)...))}
}
