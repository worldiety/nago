package xdialog

import (
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui"
	"strconv"
)

type InputBoxOptions[T any] struct {
	Title       string
	Prompt      string
	Value       *T
	OnConfirmed func()
	OnDismissed func()
}

func InputBoxText[T ~string](owner ui.ModalOwner, options InputBoxOptions[T]) {
	if options.Value == nil {
		panic("value pointer must not be nil")
	}

	owner.Modals().Append(ui.NewDialog(func(dlg *ui.Dialog) {
		var input *ui.TextField
		dlg.Title().Set(options.Title)
		dlg.Size().Set(ora.ElementSizeMedium)
		dlg.Body().Set(ui.NewVStack(func(vstack *ui.FlexContainer) {
			vstack.Append(
				ui.MakeText(options.Prompt),
				ui.NewTextField(func(textField *ui.TextField) {
					input = textField
					textField.Value().Set(string(*options.Value))
				}),
			)
		}))

		var buttons []*ui.Button

		if options.OnDismissed != nil {
			buttons = append(buttons, ui.NewActionButton("Abbrechen", func() {
				options.OnDismissed()
				owner.Modals().Remove(dlg)
			}))
		}

		if options.OnConfirmed != nil {
			buttons = append(buttons, ui.NewActionButton("Übernehmen", func() {
				*options.Value = T(input.Value().Get())
				options.OnConfirmed()
				owner.Modals().Remove(dlg)
			}))
		}

		dlg.Footer().Set(Footer(buttons...))
	}))
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

func InputBoxInt[T Number](owner ui.ModalOwner, options InputBoxOptions[T]) {
	if options.Value == nil {
		panic("value pointer must not be nil")
	}

	owner.Modals().Append(ui.NewDialog(func(dlg *ui.Dialog) {
		var input *ui.NumberField
		dlg.Title().Set(options.Title)
		dlg.Size().Set(ora.ElementSizeMedium)
		dlg.Body().Set(ui.NewVStack(func(vstack *ui.FlexContainer) {
			vstack.Append(
				ui.MakeText(options.Prompt),
				ui.NewNumberField(func(textField *ui.NumberField) {
					input = textField
					textField.Value().Set(strconv.Itoa(int(*options.Value)))
				}),
			)
		}))

		var buttons []*ui.Button

		if options.OnDismissed != nil {
			buttons = append(buttons, ui.NewActionButton("Abbrechen", func() {
				options.OnDismissed()
				owner.Modals().Remove(dlg)
			}))
		}

		if options.OnConfirmed != nil {
			buttons = append(buttons, ui.NewActionButton("Übernehmen", func() {
				iv, _ := strconv.ParseInt(input.Value().Get(), 10, 64)
				*options.Value = T(iv)
				options.OnConfirmed()
				owner.Modals().Remove(dlg)
			}))
		}

		dlg.Footer().Set(Footer(buttons...))
	}))
}
