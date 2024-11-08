package alert

import (
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	ui "go.wdy.de/nago/presentation/ui"
)

type Option interface {
	apply(opts *alertOpts)
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
}

type optFunc func(opts *alertOpts)

func (f optFunc) apply(opts *alertOpts) {
	f(opts)
}

func Ok() Option {
	return optFunc(func(opts *alertOpts) {
		opts.okBtn = ui.PrimaryButton(func() {
			opts.state.Set(false)
		}).Title("Schließen")
	})
}

func Delete(onDelete func()) Option {
	return optFunc(func(opts *alertOpts) {
		opts.delBtn = ui.PrimaryButton(func() {
			opts.state.Set(false)
			if onDelete != nil {
				onDelete()
			}
		}).Title("Löschen")
	})
}

func Save(onSave func() (close bool)) Option {
	return optFunc(func(opts *alertOpts) {
		opts.saveBtn = ui.PrimaryButton(func() {
			open := false
			if onSave != nil {
				open = !onSave()
			}

			opts.state.Set(open)

		}).Title("Speichern")
	})
}

func Closeable() Option {
	return optFunc(func(opts *alertOpts) {
		opts.closeable = ui.TertiaryButton(func() {
			opts.state.Set(false)
		}).PreIcon(heroOutline.XMark)
	})
}

// Custom adds a custom footer (button) element.
func Custom(makeCustomView func(close func(closeDlg bool)) core.View) Option {
	return optFunc(func(opts *alertOpts) {
		opts.custom = append(opts.custom, makeCustomView(func(closeDlg bool) {
			opts.state.Set(!closeDlg)
		}))
	})
}

func Cancel(onCancel func()) Option {
	return optFunc(func(opts *alertOpts) {
		opts.cancelBtn = ui.SecondaryButton(func() {
			opts.state.Set(false)
			if onCancel != nil {
				onCancel()
			}
		}).Title("Abbrechen")
	})
}

func Alignment(alignment ui.Alignment) Option {
	return optFunc(func(opts *alertOpts) {
		opts.dlgAlign = alignment
	})
}

func ModalPadding(padding ui.Padding) Option {
	return optFunc(func(opts *alertOpts) {
		opts.modalPadding = padding
	})
}

func Dialog(title string, body core.View, isPresented *core.State[bool], opts ...Option) core.View {
	var options alertOpts
	options.state = isPresented
	for _, opt := range opts {
		opt.apply(&options)
	}

	return ui.If(isPresented.Get(), ui.Modal(
		//Alignment(ui.Leading).Frame(ui.Frame{}.FullWidth())
		ui.With(ui.Dialog(ui.ScrollView(body).Frame(ui.Frame{MaxHeight: "calc(100dvh - 12rem)"}.FullWidth())).
			Title(ui.If(title != "", ui.Text(title))), func(dialog ui.TDialog) ui.TDialog {
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
				TitleX(options.closeable).
				Alignment(options.dlgAlign).
				ModalPadding(options.modalPadding)

			if len(btns) > 0 {
				dialog = dialog.Footer(ui.HStack(btns...).Gap(ui.L8))
			}
			return dialog
		})),
	)
}
