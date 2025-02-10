package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/picker"
	"slices"
	"strings"
)

type PickOneStyle int

const (
	PickOneStyleWithPicker PickOneStyle = iota
	PickOneStyleWithRadioButton
)

type PickOneOptions[T any] struct {
	Label          string
	Values         []T
	Style          PickOneStyle // Default is PickOneStyleWithPicker
	ItemRenderer   func(T) core.View
	SupportingText string
}

// PickOne binds a single field of an arbitrary value type to an associate picker. To pick an entity based
// on a foreign key semantics, use [OneToOne].
func PickOne[E any, T comparable](opts PickOneOptions[T], property Property[E, std.Option[T]]) Field[E] {
	if opts.ItemRenderer == nil {
		opts.ItemRenderer = func(t T) core.View {
			return ui.Text(fmt.Sprintf("%v", t))
		}
	}

	return Field[E]{
		Label:          opts.Label,
		SupportingText: opts.SupportingText,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[[]T](self.Window, self.ID+"-form.local").Init(func() []T {
				var tmp E
				tmp = entity.Get()
				optT := property.Get(&tmp)
				if optT.IsSome() {
					return []T{optT.Unwrap()}
				}

				return []T{}
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue []T) {
				var tmp E
				tmp = entity.Get()
				if len(newValue) == 0 {
					property.Set(&tmp, std.None[T]())
				} else {
					property.Set(&tmp, std.Some[T](newValue[0]))
				}

				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			if self.requiresValidation() {
				state.Notify()
			}

			// radiobutton picker
			if opts.Style == PickOneStyleWithRadioButton {
				stateGroup := ui.AutoRadioStateGroup(self.Window, self.ID+"-pick-one-crud-field", len(opts.Values))
				stateGroup.Observe(func(idx int) {
					if idx == -1 {
						state.Set(nil)
					} else {
						state.Set([]T{opts.Values[idx]})
					}
					state.Notify()
				})
				if stateGroup.SelectedIndex() == -1 && len(state.Get()) > 0 {
					idx := slices.Index(opts.Values, state.Get()[0])
					if idx > -1 {
						stateGroup.SetSelectedIndex(idx)

					}

				}

				views := ui.Each2(stateGroup.All(), func(idx int, checked *core.State[bool]) core.View {
					return ui.HStack(
						ui.RadioButton(checked.Get()).
							InputChecked(checked),
						ui.VStack(opts.ItemRenderer(opts.Values[idx])).
							Action(func() {
								stateGroup.SetSelectedIndex(idx)
								stateGroup.Notify()
							}),
					)
				})

				if errState.Get() != "" {
					views = append(views, ui.Text(errState.Get()))
				} else if self.SupportingText != "" {
					views = append(views, ui.Text(self.SupportingText))
				}

				return ui.VStack(views...).
					Alignment(ui.Leading).
					Gap(ui.L8).
					Frame(ui.Frame{}.FullWidth())
			}

			// default
			return picker.Picker[T](opts.Label, opts.Values, state).
				Title(self.Label).
				ItemRenderer(opts.ItemRenderer).
				MultiSelect(false).
				Disabled(self.Disabled).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := property.Get(&tmp)
			if v.IsSome() {
				return ui.TableCell(opts.ItemRenderer(v.Unwrap()))
			}
			return ui.TableCell(ui.Text(fmtOptOne(v)))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := property.Get(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(fmtOptOne(v)),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := property.Get(&a)
			bv := property.Get(&b)
			return strings.Compare(fmtOptOne(av), fmtOptOne(bv))
		},
		Stringer: func(e E) string {
			return fmtOptOne(property.Get(&e))
		},
	}
}

func fmtOptOne[T any](v std.Option[T]) string {
	if v.IsSome() {
		return fmt.Sprintf("%v", v.Unwrap())
	}

	return ""
}
