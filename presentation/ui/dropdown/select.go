// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dropdown

import (
	"fmt"

	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/internal"
)

// Option represents a selectable item in a dropdown (if enabled). Note that by definition this simple dropdown
// will never support multiple selections nor customizable views. The intention is to directly map to a native
// frontend control and require the lowest common denominator, which is the select element in the web frontend.
// UX-wise multiple selection would be possible, but the native web layout is so awkward to use and view that
// it does not make sense to support it.
type Option[ID ~string] struct {
	Value    ID
	Label    string
	Disabled bool
}

// TDropdown is a basic component (Select).
// It allows users to select one option of a given set of options,
type TDropdown[ID ~string] struct {
	label          string            // label displayed above or inside the field
	value          ID                // current value
	inputValue     *core.State[ID]   // optional external binding for controlled state
	options        []Option[ID]      // list of options to choose from
	supportingText string            // helper text shown below the field
	errorText      string            // error message shown below the field
	disabled       bool              // when true, interaction is disabled
	leading        core.View         // optional leading element (e.g., icon)
	style          ui.TextFieldStyle // visual style of the field (outlined, filled, etc.)
	frame          ui.Frame          // layout constraints
	id             string            // unique identifier for the select
	autocomplete   string            // autocomplete tags of the select
}

// Dropdown represents a user interface element which lets the user select one option from a list.
func Dropdown[ID ~string](label string, options []Option[ID], value ID) TDropdown[ID] {
	c := TDropdown[ID]{
		options: options,
		value:   value,
		label:   label,
	}

	return c
}

// FromSlice mimics the default signature of the [picker.Picker] factory so that TDropdown can be used as a drop-in
// replacement for single selection.
func FromSlice[T data.Aggregate[ID], ID ~string](label string, values []T, selectedState *core.State[[]T]) TDropdown[ID] {
	opts := make([]Option[ID], 0, len(values))
	for _, value := range values {
		opts = append(opts, Option[ID]{
			Value: value.Identity(),
			Label: fmt.Sprintf("%v", value),
		})
	}

	selectedValue := core.DerivedState[ID](selectedState, "-val").Init(func() ID {
		if sel := selectedState.Get(); len(sel) > 0 {
			return sel[0].Identity()
		}

		return ""
	})

	selectedValue.Observe(func(newValue ID) {
		if newValue == "" {
			selectedState.Update([]T{})
			return
		}

		var slice []T
		for _, value := range values {
			if value.Identity() == newValue {
				slice = append(slice, value)
				break
			}
		}

		selectedState.Update(slice)
	})

	selectedState.Observe(func(newValue []T) {
		if len(newValue) == 0 {
			selectedValue.Update("")
		} else {
			selectedValue.Update(newValue[0].Identity())
		}

	})

	return TDropdown[ID]{
		options: opts,
		value:   selectedValue.Get(),
		label:   label,
	}.InputValue(selectedValue)

}

// Label sets the label displayed above or inside the select.
func (c TDropdown[ID]) Label(label string) TDropdown[ID] {
	c.label = label
	return c
}

// Value sets the initial value of the select.
func (c TDropdown[ID]) Value(value ID) TDropdown[ID] {
	c.value = value
	return c
}

// InputValue binds the select to an external value state,
// allowing it to be controlled from outside the component.
func (c TDropdown[ID]) InputValue(input *core.State[ID]) TDropdown[ID] {
	c.inputValue = input
	return c
}

// Options sets the list of options available for selection.
func (c TDropdown[ID]) Options(options []Option[ID]) TDropdown[ID] {
	c.options = options
	return c
}

// SupportingText sets the supporting text displayed below the select.
func (c TDropdown[ID]) SupportingText(text string) TDropdown[ID] {
	c.supportingText = text
	return c
}

// ErrorText sets the error text displayed below the select.
func (c TDropdown[ID]) ErrorText(text string) TDropdown[ID] {
	c.errorText = text
	return c
}

// Disabled enables or disables user interaction with the select.
func (c TDropdown[ID]) Disabled(disabled bool) TDropdown[ID] {
	c.disabled = disabled
	return c
}

// Leading sets a leading view for the select.
// This view is displayed at the start of the select, e.g., an icon.
func (c TDropdown[ID]) Leading(v core.View) TDropdown[ID] {
	c.leading = v
	return c
}

// Style sets the visual style of the select.
func (c TDropdown[ID]) Style(s ui.TextFieldStyle) TDropdown[ID] {
	c.style = s
	return c
}

// Frame sets the layout frame of the field (size, width, height, etc.).
func (c TDropdown[ID]) Frame(frame ui.Frame) TDropdown[ID] {
	c.frame = frame
	return c
}

// ID assigns a unique identifier to the select, useful for testing or referencing.
func (c TDropdown[ID]) ID(id string) TDropdown[ID] {
	c.id = id
	return c
}

// Autocomplete defines the autocomplete tags of the input
func (c TDropdown[ID]) Autocomplete(tags string) TDropdown[ID] {
	c.autocomplete = tags
	return c
}

// Render builds and returns the protocol representation of the select.
func (c TDropdown[ID]) Render(ctx core.RenderContext) core.RenderNode {
	options := make([]proto.SelectOption, 0, len(c.options))
	for _, option := range c.options {
		options = append(options, proto.SelectOption{
			Value:    proto.Str(option.Value),
			Disabled: proto.Bool(option.Disabled),
			Label:    proto.Str(option.Label),
		})
	}

	return &proto.Select{
		Label:          proto.Str(c.label),
		SupportingText: proto.Str(c.supportingText),
		ErrorText:      proto.Str(c.errorText),
		Value:          proto.Str(c.value),
		Frame:          internal.FrameToOra(c.frame),
		InputValue:     c.inputValue.Ptr(),
		Style:          proto.TextFieldStyle(c.style),
		Leading:        internal.Render(ctx, c.leading),
		Disabled:       proto.Bool(c.disabled),
		Id:             proto.Str(c.id),
		Options:        options,
		Autocomplete:   proto.Str(c.autocomplete),
	}
}
