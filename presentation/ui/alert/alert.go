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

func Ok() Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.okBtn = ui.PrimaryButton(func() {
			opts.state.Set(false)
			opts.state.Notify()
		}).Title(rstring.ActionClose.Get(wnd))
	})
}

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

func Save(onSave func() (close bool)) Option {
	return save(rstring.ActionSave, onSave)
}

func Create(onSave func() (close bool)) Option {
	return save(rstring.ActionCreate, onSave)
}

func Apply(onSave func() (close bool)) Option {
	return save(rstring.ActionApply, onSave)
}

func Close(onSave func() (close bool)) Option {
	return save(rstring.ActionClose, onSave)
}

func Confirm(onSave func() (close bool)) Option {
	return save(rstring.ActionConfirm, onSave)
}

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

func Width(w ui.Length) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.width = w
	})
}

func Large() Option {
	return Width(ui.L560)
}

func Larger() Option {
	return Width(ui.L880)
}

func XLarge() Option {
	return Width(ui.L1200)
}

func XXLarge() Option {
	return Width(ui.L1600)
}

func FullHeight() Option {
	return Height(ui.Full)
}

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

func Alignment(alignment ui.Alignment) Option {
	return optFunc(func(wnd core.Window, opts *alertOpts) {
		opts.dlgAlign = alignment
	})
}

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

// TDialog is an overlay component(Dialog).
type TDialog struct {
	title       string
	body        core.View
	isPresented *core.State[bool]
	opts        []Option
}

func Dialog(title string, body core.View, isPresented *core.State[bool], opts ...Option) TDialog {
	return TDialog{
		title:       title,
		body:        body,
		isPresented: isPresented,
		opts:        opts,
	}
}

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

	return ui.Modal(
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
		})).Render(ctx)
}
