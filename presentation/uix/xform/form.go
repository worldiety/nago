package xform

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xdialog"
	"time"
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
	~float32 | ~float64
}

type MapF[From, To any] func(From) To

type Options struct {
}

type Field struct {
	Label    string
	Group    GroupID // Group reference, if not exist, it is used as the label
	Hint     string
	Disabled bool
	Secure   bool
}

type GroupID string

type Group struct {
	ID    GroupID
	Label string
}

type Binding struct {
	elems     []formElem
	Groups    []Group // defines the order and settings of form groups, e.g. if collapsed etc.
	OnChanged func()
}

func NewBinding() *Binding {
	return &Binding{}
}

func Slider[T Number](binding *Binding, target *T, minIncl, maxIncl, stepSize T, opts Field) {
	tf := ui.NewSlider(nil)
	tf.Label().Set(opts.Label)
	tf.Min().Set(float64(*target))
	tf.Hint().Set(opts.Hint)
	tf.Disabled().Set(opts.Disabled)
	tf.Min().Set(float64(minIncl))
	tf.Max().Set(float64(maxIncl))
	tf.Stepsize().Set(float64(stepSize))
	tf.OnChanged().Set(func() {
		*target = T(tf.Min().Get())
		if binding.OnChanged != nil {
			binding.OnChanged()
		}
	})
	binding.elems = append(binding.elems, formElem{
		getComponent: func() core.Component {
			return tf
		},
		opts: opts,
	})
}

func Int[T string](binding *Binding, target *T, opts Field) {
	tf := ui.NewNumberField(nil)
	tf.Label().Set(opts.Label)
	tf.Value().Set(string(*target))
	tf.Hint().Set(opts.Hint)
	tf.Disabled().Set(opts.Disabled)
	tf.OnValueChanged().Set(func() {
		*target = T(tf.Value().Get())
		if binding.OnChanged != nil {
			binding.OnChanged()
		}
	})
	binding.elems = append(binding.elems, formElem{
		getComponent: func() core.Component {
			return tf
		},
		opts: opts,
	})
}

// Date binds the given time. Please check your time.Location. This depends on the host and may be anything,
// depending on the host where this process is executed.
func Date(binding *Binding, target *time.Time, opts Field) {
	if target.IsZero() {
		*target = time.Now() // TODO don't know, but the zero value is unusable
	}
	tz := target.Location()
	tf := ui.NewDatepicker(nil)
	tf.Label().Set(opts.Label)
	tme := *target
	tf.SelectedStartYear().Set(int64(tme.Year()))
	tf.SelectedStartDay().Set(int64(tme.Day()))
	tf.SelectedStartMonth().Set(int64(tme.Month()))
	tf.Hint().Set(opts.Hint)
	tf.Disabled().Set(opts.Disabled)
	tf.OnClicked().Set(func() {
		tf.Expanded().Set(!tf.Expanded().Get())
	})
	tf.OnSelectionChanged().Set(func() {
		newTime := time.Date(int(tf.SelectedStartYear().Get()), time.Month(tf.SelectedStartMonth().Get()+1), int(tf.SelectedStartDay().Get()), 0, 0, 0, 0, tz)
		fmt.Println(newTime)
		*target = newTime
		if binding.OnChanged != nil {
			binding.OnChanged()
		}
	})

	binding.elems = append(binding.elems, formElem{
		getComponent: func() core.Component {
			return tf
		},
		opts: opts,
	})
}

func String[T ~string](binding *Binding, target *T, opts Field) {
	tf := ui.NewTextField(nil)
	tf.Label().Set(opts.Label)
	tf.Value().Set(string(*target))
	tf.Hint().Set(opts.Hint)
	tf.Disabled().Set(opts.Disabled)
	tf.OnTextChanged().Set(func() {
		*target = T(tf.Value().Get())
		if binding.OnChanged != nil {
			binding.OnChanged()
		}
	})
	binding.elems = append(binding.elems, formElem{
		getComponent: func() core.Component {
			return tf
		},
		opts: opts,
	})
}

func Bool[T ~bool](binding *Binding, target *T, opts Field) {
	tf := ui.NewToggle(nil)
	tf.Label().Set(opts.Label)
	tf.Checked().Set(bool(*target))
	// tf.Hint().Set(opts.Hint) // TODO hint is missing
	tf.Disabled().Set(opts.Disabled)
	tf.OnCheckedChanged().Set(func() {
		*target = T(tf.Checked().Get())
		if binding.OnChanged != nil {
			binding.OnChanged()
		}
	})
	binding.elems = append(binding.elems, formElem{
		getComponent: func() core.Component {
			return tf
		},
		opts: opts,
	})
}

func OneToOne[E data.Aggregate[ID], ID data.IDType](binding *Binding, target *ID, items iter.Seq2[E, error], itemCaptionizer MapF[E, string], opts Field) {
	cb := ui.NewDropdown(nil)
	cb.Label().Set(opts.Label)
	cb.Hint().Set(opts.Hint)
	cb.OnClicked().Set(func() {
		cb.Expanded().Set(!cb.Expanded().Get())
	})
	cb.Multiselect().Set(false)
	cb.Disabled().Set(opts.Disabled)
	cb.Hint().Set(opts.Hint)

	var err error
	itemSlice := slices.Collect(iter.BreakOnError(&err, items))
	if err != nil {
		binding.elems = append(binding.elems, formElem{
			func() core.Component {
				return xdialog.ErrorView("cannot collect dropdown items", err)
			}, opts})

		return
	}

	var zero ID
	for i, item := range itemSlice {
		isSelected := *target == item.Identity()

		if isSelected {
			cb.SelectedIndices().Append(int64(i))
		}

		cb.Items().Append(
			ui.NewDropdownItem(func(dropdownItem *ui.DropdownItem) {
				dropdownItem.Content().Set(itemCaptionizer(item))
				dropdownItem.OnClicked().Set(func() {
					cb.Toggle(dropdownItem)
					*target = zero

					cb.SelectedIndices().Iter(func(i int64) bool {
						*target = itemSlice[i].Identity()
						return true
					})

					if binding.OnChanged != nil {
						binding.OnChanged()
					}
				})
			}),
		)
	}

	binding.elems = append(binding.elems, formElem{func() core.Component {
		return cb
	}, opts})
}

func OneToMany[Slice ~[]ID, E data.Aggregate[ID], ID data.IDType](binding *Binding, target *Slice, items iter.Seq2[E, error], itemCaptionizer MapF[E, string], opts Field) {
	cb := ui.NewDropdown(nil)
	cb.Label().Set(opts.Label)
	cb.Hint().Set(opts.Hint)
	cb.OnClicked().Set(func() {
		cb.Expanded().Set(!cb.Expanded().Get())
	})
	cb.Multiselect().Set(true)
	cb.Disabled().Set(opts.Disabled)
	cb.Hint().Set(opts.Hint)

	var err error
	itemSlice := slices.Collect(iter.BreakOnError(&err, items))
	if err != nil {
		binding.elems = append(binding.elems, formElem{
			func() core.Component {
				return xdialog.ErrorView("cannot collect dropdown items", err)
			}, opts})

		return
	}

	for i, item := range itemSlice {
		isSelected := false
		for _, id := range *target {
			if id == item.Identity() {
				isSelected = true
				break
			}
		}

		if isSelected {
			cb.SelectedIndices().Append(int64(i))
		}

		cb.Items().Append(
			ui.NewDropdownItem(func(dropdownItem *ui.DropdownItem) {
				dropdownItem.Content().Set(itemCaptionizer(item))
				dropdownItem.OnClicked().Set(func() {
					cb.Toggle(dropdownItem)
					*target = nil

					cb.SelectedIndices().Iter(func(i int64) bool {
						*target = append(*target, itemSlice[i].Identity())
						return true
					})

					if binding.OnChanged != nil {
						binding.OnChanged()
					}
				})
			}),
		)
	}

	binding.elems = append(binding.elems, formElem{func() core.Component {
		return cb
	}, opts})
}

// NewForm creates a form, based on the given binding.
func NewForm(binding *Binding) core.Component {
	type group struct {
		definedGroup Group
		elems        []formElem
	}

	var groups []*group

	// what the dev wants
	for _, g := range binding.Groups {
		groups = append(groups, &group{
			definedGroup: g,
		})
	}

	// add the anon group
	groups = append(groups, &group{})

	// add unknown groups
	for _, elem := range binding.elems {
		found := false
		for _, g := range groups {
			if g.definedGroup.ID == elem.opts.Group {
				found = true
				break
			}
		}

		if !found {
			groups = append(groups, &group{
				definedGroup: Group{
					ID:    elem.opts.Group,
					Label: string(elem.opts.Group),
				},
			})
		}

	}

	// order fields according to groups, no groups at all, will result in a single unnamed group
nextElem:
	for _, elem := range binding.elems {
		for _, g := range groups {
			if g.definedGroup.ID == elem.opts.Group {
				g.elems = append(g.elems, elem)
				continue nextElem
			}
		}
	}

	return ui.NewVBox(func(vbox *ui.VBox) {
		for i, g := range groups {
			if len(g.elems) == 0 {
				continue // do not show empty sections
			}

			if g.definedGroup.ID != "" {
				// only add a section header for defined groups
				vbox.Append(ui.NewText(func(text *ui.Text) {
					text.Size().Set("xl")
					text.Value().Set(g.definedGroup.Label)
				}))
			}

			for _, elem := range g.elems {
				vbox.Append(elem.getComponent())
			}

			if i < len(groups)-2 {
				vbox.Append(ui.NewDivider(nil))
			}
		}

	})
}

type formElem struct {
	getComponent func() core.Component
	opts         Field
}

func Show(modals ui.ModalOwner, binding *Binding, onSave func() error) {
	modals.Modals().Append(
		ui.NewDialog(func(dlg *ui.Dialog) {
			dlg.Actions().Append(
				ui.NewButton(func(btn *ui.Button) {
					btn.Caption().Set("Speichern")
					btn.Action().Set(func() {
						if xdialog.HandleError(modals, "cannot save item", onSave()) {
							return
						}
						modals.Modals().Remove(dlg)
					})
				}),
				ui.NewButton(func(btn *ui.Button) {
					btn.Caption().Set("Abbrechen")
					btn.Action().Set(func() {
						modals.Modals().Remove(dlg)
					})
				}),
			)
			dlg.Body().Set(NewForm(binding))
		}),
	)
}
