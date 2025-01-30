package ui

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"iter"
)

type RadioStateGroup struct {
	states []*core.State[bool]
}

func AutoRadioStateGroup(wnd core.Window, id string, states int) RadioStateGroup {
	var wndStates = make([]*core.State[bool], 0, states)
	for i := range states {
		state := core.StateOf[bool](wnd, fmt.Sprintf("%s-%d", id, i))
		state.Observe(func(newValue bool) {
			for i2, wndState := range wndStates {
				wndState.Set(i == i2)
			}
		})
		wndStates = append(wndStates, state)
	}

	return RadioStateGroup{states: wndStates}
}

func (s RadioStateGroup) Observe(f func(newIdx int)) {
	for _, state := range s.states {
		state.Observe(func(newValue bool) {
			f(s.SelectedIndex())
		})
	}
}

func (s RadioStateGroup) SetSelectedIndex(idx int) {
	for i, state := range s.states {
		state.Set(idx == i)
	}
}

func (s RadioStateGroup) Notify() {
	for _, state := range s.states {
		state.Notify()
	}
}

// SelectedIndex returns -1 or the selected index.
func (s RadioStateGroup) SelectedIndex() int {
	for i, state := range s.states {
		if state.Get() {
			return i
		}
	}

	return -1
}

func (s RadioStateGroup) All() iter.Seq2[int, *core.State[bool]] {
	return func(yield func(int, *core.State[bool]) bool) {
		for i := range s.states {
			if !yield(i, s.states[i]) {
				return
			}
		}
	}
}

type TRadioButton struct {
	value      bool
	inputValue *core.State[bool]
	disabled   bool
	invisible  bool
}

// RadioButton represents a user interface element which spans a visible area to click or tap from the user.
// Use it for controls, which do not cause an immediate effect and only one element can be picked at a time.
// See also [Toggle], [Checkbox] and [Select].
func RadioButton(checked bool) TRadioButton {
	c := TRadioButton{
		value: checked,
	}

	return c
}

func (c TRadioButton) InputChecked(input *core.State[bool]) TRadioButton {
	c.inputValue = input
	return c
}

func (c TRadioButton) Disabled(disabled bool) TRadioButton {
	c.disabled = disabled
	return c
}

func (c TRadioButton) Visible(v bool) TRadioButton {
	c.invisible = !v
	return c
}

func (c TRadioButton) Render(ctx core.RenderContext) core.RenderNode {

	return &proto.Radiobutton{
		Value:      proto.Bool(c.value),
		InputValue: c.inputValue.Ptr(),
		Disabled:   proto.Bool(c.disabled),
		Invisible:  proto.Bool(c.invisible),
	}
}
