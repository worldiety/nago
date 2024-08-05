package alert

import (
	"go.wdy.de/nago/presentation/core"
	ui "go.wdy.de/nago/presentation/ui2"
)

type Option interface {
	apply(opts *alertOpts)
}

type alertOpts struct {
	state *core.State[bool]
	okBtn core.View
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

func Dialog(title, message string, isPresented *core.State[bool], opts ...Option) core.View {
	var options alertOpts
	options.state = isPresented
	for _, opt := range opts {
		opt.apply(&options)
	}

	return ui.If(isPresented.Get(), ui.Modal(
		ui.With(ui.Dialog(ui.VStack(ui.Text(message)).Alignment(ui.Leading).Frame(ui.Frame{}.FullWidth())).
			Title(ui.Text(title)), func(dialog ui.TDialog) ui.TDialog {
			var btns []core.View
			if options.okBtn != nil {
				btns = append(btns, options.okBtn)
			}

			dialog = dialog.Footer(ui.HStack(btns...))
			return dialog
		})),
	)
}
