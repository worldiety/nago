// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package accessibility

import (
	"strconv"
	"strings"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// AriaRole is a WAI-ARIA role, which tells assistive technologies like screen readers what kind of
// element a component represents. See https://www.w3.org/TR/wai-aria/#role_definitions.
// Any valid role can be used, the declared constants are just the most common ones.
type AriaRole string

const (
	RoleAlert        AriaRole = "alert"
	RoleAlertDialog  AriaRole = "alertdialog"
	RoleArticle      AriaRole = "article"
	RoleBanner       AriaRole = "banner"
	RoleButton       AriaRole = "button"
	RoleCheckbox     AriaRole = "checkbox"
	RoleComplement   AriaRole = "complementary"
	RoleContentInfo  AriaRole = "contentinfo"
	RoleDialog       AriaRole = "dialog"
	RoleFigure       AriaRole = "figure"
	RoleGroup        AriaRole = "group"
	RoleHeading      AriaRole = "heading"
	RoleImg          AriaRole = "img"
	RoleLink         AriaRole = "link"
	RoleList         AriaRole = "list"
	RoleListItem     AriaRole = "listitem"
	RoleMain         AriaRole = "main"
	RoleMenu         AriaRole = "menu"
	RoleMenuItem     AriaRole = "menuitem"
	RoleNavigation   AriaRole = "navigation"
	RoleNone         AriaRole = "none"
	RolePresentation AriaRole = "presentation"
	RoleProgressBar  AriaRole = "progressbar"
	RoleRadio        AriaRole = "radio"
	RoleRadioGroup   AriaRole = "radiogroup"
	RoleRegion       AriaRole = "region"
	RoleSearch       AriaRole = "search"
	RoleSeparator    AriaRole = "separator"
	RoleSlider       AriaRole = "slider"
	RoleStatus       AriaRole = "status"
	RoleSwitch       AriaRole = "switch"
	RoleTab          AriaRole = "tab"
	RoleTabList      AriaRole = "tablist"
	RoleTabPanel     AriaRole = "tabpanel"
	RoleTimer        AriaRole = "timer"
	RoleToolbar      AriaRole = "toolbar"
	RoleTooltip      AriaRole = "tooltip"
)

// TAccessible is a utility component (Accessible).
// It wraps an arbitrary view and annotates it with WAI-ARIA attributes like a role, aria-hidden or
// aria-valuetext, so that assistive technologies like screen readers can understand the content.
// The attributes are applied directly to the content, no additional element is introduced.
type TAccessible struct {
	content    core.View
	role       AriaRole
	label      string
	labelledBy []string
	hidden     bool
	valueText  string
	valueNow   string
	valueMin   string
	valueMax   string
	expanded   string
	selected   string
	disabled   string
}

// Accessible wraps the given content to annotate it with WAI-ARIA attributes.
func Accessible(content core.View) TAccessible {
	return TAccessible{content: content}
}

// Role sets the WAI-ARIA role.
func (c TAccessible) Role(role AriaRole) TAccessible {
	c.role = role
	return c
}

// Label sets the aria-label, which defines the accessible name.
func (c TAccessible) Label(label string) TAccessible {
	c.label = label
	return c
}

// LabelledBy sets the aria-labelledby to the ids of the components which label this component.
func (c TAccessible) LabelledBy(ids ...string) TAccessible {
	c.labelledBy = ids
	return c
}

// Hidden sets the aria-hidden, which removes the content from the accessibility tree,
// e.g. for decorative content. The content is still visible.
func (c TAccessible) Hidden(hidden bool) TAccessible {
	c.hidden = hidden
	return c
}

// ValueText sets the aria-valuetext, which is a human readable alternative of the current value of
// a range widget, e.g. "3 of 5 steps" or "medium".
func (c TAccessible) ValueText(text string) TAccessible {
	c.valueText = text
	return c
}

// ValueNow sets the aria-valuenow, which is the current value of a range widget.
func (c TAccessible) ValueNow(v float64) TAccessible {
	c.valueNow = formatAriaFloat(v)
	return c
}

// ValueRange sets the aria-valuemin and aria-valuemax of a range widget.
func (c TAccessible) ValueRange(minValue, maxValue float64) TAccessible {
	c.valueMin = formatAriaFloat(minValue)
	c.valueMax = formatAriaFloat(maxValue)
	return c
}

// Expanded sets the aria-expanded, which tells whether the controlled content is expanded or collapsed.
func (c TAccessible) Expanded(expanded bool) TAccessible {
	c.expanded = strconv.FormatBool(expanded)
	return c
}

// Selected sets the aria-selected, e.g. of a tab or an option.
func (c TAccessible) Selected(selected bool) TAccessible {
	c.selected = strconv.FormatBool(selected)
	return c
}

// Disabled sets the aria-disabled, which tells that the content is perceivable but not operable.
func (c TAccessible) Disabled(disabled bool) TAccessible {
	c.disabled = strconv.FormatBool(disabled)
	return c
}

// Render builds and returns the protocol representation of the accessible wrapper.
func (c TAccessible) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.Accessible{
		Content:    c.content.Render(ctx),
		Role:       proto.Str(c.role),
		Label:      proto.Str(c.label),
		LabelledBy: proto.Str(strings.Join(c.labelledBy, " ")),
		Hidden:     proto.Bool(c.hidden),
		ValueText:  proto.Str(c.valueText),
		ValueNow:   proto.Str(c.valueNow),
		ValueMin:   proto.Str(c.valueMin),
		ValueMax:   proto.Str(c.valueMax),
		Expanded:   proto.Str(c.expanded),
		Selected:   proto.Str(c.selected),
		Disabled:   proto.Str(c.disabled),
	}
}

func formatAriaFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
