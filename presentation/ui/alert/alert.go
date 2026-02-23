// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package alert

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	ui "go.wdy.de/nago/presentation/ui"
)

type Option interface {
	apply(wnd core.Window, opts *alertOpts)
}

type alertOpts struct {
	state        *core.State[bool]
	okBtn        core.View
	delBtn       core.View
	saveBtn      core.View
	cancelBtn    core.View
	custom       []core.View
	closeable    core.View
	dlgAlign     ui.Alignment
	modalPadding ui.Padding
	preBody      core.View
	minWidth     ui.Length
	width        ui.Length
	height       ui.Length
}

type optFunc func(wnd core.Window, opts *alertOpts)

func (f optFunc) apply(wnd core.Window, opts *alertOpts) {
	f(wnd, opts)
}

// Ok adds a default button that closes the dialog.
func Ok() Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.okBtn = ui.PrimaryButton(func() {
			opts.state.Set(false)
			opts.state.Notify()
		}).Title(rstring.ActionClose.Get(wnd))
	})
}

// Delete adds a button that closes the dialog and triggers the given callback.
func Delete(onDelete func()) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.delBtn = ui.PrimaryButton(func() {
			opts.state.Set(false)
			opts.state.Notify()
			if onDelete != nil {
				onDelete()
			}
		}).Title(rstring.ActionDelete.Get(wnd))
	})
}

// Save adds a button that triggers the callback and optionally closes the dialog.
func Save(onSave func() (close bool)) Option {
	return save(rstring.ActionSave, onSave)
}

func Create(onSave func() (close bool)) Option {
	return save(rstring.ActionCreate, onSave)
}

func Add(onSave func() (close bool)) Option {
	return save(rstring.ActionAdd, onSave)
}

// Apply adds an button that triggers the callback and optionally closes the dialog.
func Apply(onSave func() (close bool)) Option {
	return save(rstring.ActionApply, onSave)
}

// Close adds a button that triggers the callback and optionally closes the dialog.
func Close(onSave func() (close bool)) Option {
	return save(rstring.ActionClose, onSave)
}

// Confirm adds a button that triggers the callback and optionally closes the dialog.
func Confirm(onSave func() (close bool)) Option {
	return save(rstring.ActionConfirm, onSave)
}

// save is a helper that creates a button with the given caption and onSave logic.
func save(caption i18n.StrHnd, onSave func() (close bool)) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.saveBtn = ui.PrimaryButton(func() {
			open := false
			if onSave != nil {
				open = !onSave()
			}

			opts.state.Set(open)
			opts.state.Notify()

		}).Title(caption.Get(wnd))
	})
}

// Closeable adds a close icon button ("X") to dismiss the dialog.
// It also installs a dismiss-listener on the modal to close the dialog automatically e.g. when pressing the escape key.
func Closeable() Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.closeable = ui.TertiaryButton(func() {
			opts.state.Set(false)
			opts.state.Notify()
		}).PreIcon(heroOutline.XMark)
	})
}

// MinWidth is probably not what you want. Consider [Width] instead.
func MinWidth(w ui.Length) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.minWidth = w
	})
}

// Width sets the dialog width to the given Length.
func Width(w ui.Length) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.width = w
	})
}

// Large sets the dialog width to 560dp (35rem).
func Large() Option {
	return Width(ui.L560)
}

// Larger sets the dialog width to 880dp (55rem).
func Larger() Option {
	return Width(ui.L880)
}

// XLarge sets the dialog width to 1200dp (75rem).
func XLarge() Option {
	return Width(ui.L1200)
}

// XXLarge sets the dialog width to 1600dp (100rem).
func XXLarge() Option {
	return Width(ui.L1600)
}

// FullHeight makes the dialog take the full available height.
func FullHeight() Option {
	return Height(ui.Full)
}

// Height sets the dialog height to the given Length.
func Height(h ui.Length) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.height = h
	})
}

// Custom adds a custom footer (button) element.
func Custom(makeCustomView func(close func(closeDlg bool)) core.View) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.custom = append(opts.custom, makeCustomView(func(closeDlg bool) {
			opts.state.Set(!closeDlg)
			opts.state.Notify()
		}))
	})
}

// Cancel adds an button that closes the dialog and triggers the given callback.
// It also installs a dismiss-listener on the modal to close the dialog automatically e.g. when pressing the escape key.
func Cancel(onCancel func()) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.cancelBtn = ui.SecondaryButton(func() {
			opts.state.Set(false)
			opts.state.Notify()
			if onCancel != nil {
				onCancel()
			}
		}).Title(rstring.ActionCancel.Get(wnd))
	})
}

// Back adds a button that closes the dialog and triggers the given callback.
func Back(onCancel func()) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.cancelBtn = ui.SecondaryButton(func() {
			opts.state.Set(false)
			opts.state.Notify()
			if onCancel != nil {
				onCancel()
			}
		}).Title(rstring.ActionBack.Get(wnd))
	})
}

// Alignment sets the alignment of the dialog content.
func Alignment(alignment ui.Alignment) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.dlgAlign = alignment
	})
}

// ModalPadding defines the padding inside the modal dialog.
func ModalPadding(padding ui.Padding) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.modalPadding = padding
	})
}

// PreBody sets a view between title and body. The body will scroll automatically, however the title and pre body not.
func PreBody(v core.View) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.preBody = v
	})
}

// TDialog is an overlay component (Dialog).
// This component presents content in a centered modal window
// above the main interface. It is typically used for confirmations,
// forms, or focused interactions that require the user's attention.
//
// A dialog consists of a title and body content, and can be controlled
// through an isPresented state. Additional options can be provided
// to customize its appearance and behavior (e.g., alignment, padding,
// footer actions).
type TDialog struct {
	title       string
	body        core.View
	isPresented *core.State[bool]
	opts        []Option
}

// Dialog creates a new dialog with the given title, body content, visibility state, and optional configuration.
func Dialog(title string, body core.View, isPresented *core.State[bool], opts ...Option) TDialog {
	return TDialog{
		title:       title,
		body:        body,
		isPresented: isPresented,
		opts:        opts,
	}
}

// Render builds and displays the dialog as a modal overlay if it is currently presented.
// It applies all configured options (e.g., size, alignment, padding, buttons) to customize
// the dialog's layout and behavior. The dialog content is wrapped in a scroll view, and
// footer buttons (OK, Cancel, Delete, Save, or custom) are rendered in a consistent order.
func (t TDialog) Render(ctx core.RenderContext) core.RenderNode {
	if !t.isPresented.Get() {
		return nil
	}

	var options alertOpts
	options.state = t.isPresented
	for _, opt := range t.opts {
		opt.apply(ctx.Window(), &options)
	}

	var fixHeight ui.Length
	if options.height == ui.Full {
		fixHeight = "calc(100dvh - 12rem)"
	}

	modal := ui.Modal(
		ui.With(ui.Dialog(ui.ScrollView(t.body).Frame(ui.Frame{Height: fixHeight}.FullWidth())).
			Title(ui.If(t.title != "", ui.Text(t.title))), func(dialog ui.TDialog) ui.TDialog {
			var btns []core.View
			// we do this to keep sensible order
			if options.okBtn != nil {
				btns = append(btns, options.okBtn)
			}

			if options.cancelBtn != nil {
				btns = append(btns, options.cancelBtn)
			}

			if options.delBtn != nil {
				btns = append(btns, options.delBtn)
			}

			if options.saveBtn != nil {
				btns = append(btns, options.saveBtn)
			}

			btns = append(btns, options.custom...)

			dialog = dialog.
				PreBody(options.preBody).
				TitleX(options.closeable).
				Alignment(options.dlgAlign).
				ModalPadding(options.modalPadding)

			dialog = dialog.WithFrame(func(frame ui.Frame) ui.Frame {
				if options.minWidth != "" {
					frame.MinWidth = options.minWidth
				}

				if options.width != "" {
					frame.Width = options.width
				}

				if options.height != "" {
					frame.MaxHeight = options.height
				}

				return frame
			})

			if len(btns) > 0 {
				dialog = dialog.Footer(ui.HStack(btns...).Gap(ui.L8))
			}
			return dialog
		}))

	if options.closeable != nil || options.cancelBtn != nil {
		modal = modal.OnDismissRequest(func() {
			options.state.Set(false)
			options.state.Notify()
		})
	}

	return modal.Render(ctx)
}
