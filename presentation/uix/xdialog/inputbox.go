package xdialog

import (
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/uilegacy"
	"strconv"
)

type InputBoxOptions[T any] struct {
	Title       string
	Prompt      string
	Value       *T
	OnConfirmed func()
	OnDismissed func()
}

func InputBoxText[T ~string](owner uilegacy.ModalOwner, options InputBoxOptions[T]) {
	if options.Value == nil {
		panic("value pointer must not be nil")
	}

	owner.Modals().Append(uilegacy.NewDialog(func(dlg *uilegacy.Dialog) {
		var input *uilegacy.TextField
		dlg.Title().Set(options.Title)
		dlg.Size().Set(ora.ElementSizeMedium)
		dlg.Body().Set(uilegacy.NewVStack(func(vstack *uilegacy.VStack) {
			vstack.Append(
				uilegacy.MakeText(options.Prompt),
				uilegacy.NewTextField(func(textField *uilegacy.TextField) {
					input = textField
					textField.Value().Set(string(*options.Value))
				}),
			)
		}))

		var buttons []*uilegacy.Button

		if options.OnDismissed != nil {
			buttons = append(buttons, uilegacy.NewActionButton("Abbrechen", func() {
				options.OnDismissed()
				owner.Modals().Remove(dlg)
			}))
		}

		if options.OnConfirmed != nil {
			buttons = append(buttons, uilegacy.NewActionButton("Übernehmen", func() {
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

func InputBoxInt[T Number](owner uilegacy.ModalOwner, options InputBoxOptions[T]) {
	if options.Value == nil {
		panic("value pointer must not be nil")
	}

	owner.Modals().Append(uilegacy.NewDialog(func(dlg *uilegacy.Dialog) {
		var input *uilegacy.NumberField
		dlg.Title().Set(options.Title)
		dlg.Size().Set(ora.ElementSizeMedium)
		dlg.Body().Set(uilegacy.NewVStack(func(vstack *uilegacy.VStack) {
			vstack.Append(
				uilegacy.MakeText(options.Prompt),
				uilegacy.NewNumberField(func(textField *uilegacy.NumberField) {
					input = textField
					textField.Value().Set(strconv.Itoa(int(*options.Value)))
				}),
			)
		}))

		var buttons []*uilegacy.Button

		if options.OnDismissed != nil {
			buttons = append(buttons, uilegacy.NewActionButton("Abbrechen", func() {
				options.OnDismissed()
				owner.Modals().Remove(dlg)
			}))
		}

		if options.OnConfirmed != nil {
			buttons = append(buttons, uilegacy.NewActionButton("Übernehmen", func() {
				iv, _ := strconv.ParseInt(input.Value().Get(), 10, 64)
				*options.Value = T(iv)
				options.OnConfirmed()
				owner.Modals().Remove(dlg)
			}))
		}

		dlg.Footer().Set(Footer(buttons...))
	}))
}
