// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"iter"
	"slices"
)

type Form struct {
	ID          FormID
	Root        FormView
	name        Ident
	description string
}

func NewForm(id FormID, name Ident, root FormView) *Form {
	return &Form{
		ID:   id,
		name: name,
		Root: root,
	}
}

func (f *Form) Name() Ident {
	return f.name
}

func (f *Form) Description() string {
	return f.description
}

func (f *Form) SetDescription(s string) {
	f.description = s
}

func (f *Form) Identity() FormID {
	return f.ID
}

func (f *Form) Clone() *Form {
	return &Form{
		ID:          f.ID,
		Root:        f.Root.Clone(),
		name:        f.name,
		description: f.description,
	}
}

type FormView interface {
	Identity() ViewID
	Clone() FormView
}

// FormViewGroup extens FormView with view group parent functions. A view tree must not contain cycles.
type FormViewGroup interface {
	FormView
	All() iter.Seq[FormView]
	// Insert merges the given view into the ordered children list.
	// If after is empty, the view is appended to the beginning of the group.
	Insert(view FormView, after ViewID)
	Remove(ViewID)
}

type ViewID string

var (
	_ FormView      = (*FormText)(nil)
	_ FormView      = (*FormCheckbox)(nil)
	_ FormViewGroup = (*FormVStack)(nil)
	_ FormViewGroup = (*baseViewGroup)(nil)
	_ FormViewGroup = (*FormCard)(nil)
)

type FormText struct {
	id    ViewID
	value string
	style FormTextStyle
}

func NewFormText(id ViewID, value string, style FormTextStyle) *FormText {
	return &FormText{
		id:    id,
		value: value,
		style: style,
	}
}

func (f *FormText) Clone() FormView {
	c := *f
	return &c
}

func (f *FormText) Value() string {
	return f.value
}

func (f *FormText) Style() FormTextStyle {
	return f.style
}

func (f *FormText) Identity() ViewID {
	return f.id
}

type baseViewGroup struct {
	id    ViewID
	views []FormView
}

func (b *baseViewGroup) Identity() ViewID {
	return b.id
}

func (b *baseViewGroup) Clone() FormView {
	return b.clone()
}

func (b *baseViewGroup) clone() *baseViewGroup {
	return &baseViewGroup{
		id:    b.id,
		views: slices.Clone(b.views),
	}
}

func (b *baseViewGroup) All() iter.Seq[FormView] {
	return slices.Values(b.views)
}

func (b *baseViewGroup) Insert(view FormView, after ViewID) {
	if after == "" {
		b.views = append([]FormView{view}, b.views...)
		return
	}

	tmp := make([]FormView, 0, len(b.views)+1)
	inserted := false
	for _, v := range b.views {
		tmp = append(tmp, v)
		if v.Identity() == after {
			tmp = append(tmp, view)
			inserted = true
			break
		}
	}

	if !inserted {
		tmp = append(tmp, view)
	}

	b.views = tmp
}

func (b *baseViewGroup) Remove(id ViewID) {
	b.views = slices.DeleteFunc(b.views, func(v FormView) bool { return v.Identity() == id })
}

type FormVStack struct {
	*baseViewGroup
}

func NewFormVStack(id ViewID) *FormVStack {
	return &FormVStack{baseViewGroup: &baseViewGroup{id: id}}
}

func (f *FormVStack) Clone() FormView {
	return &FormVStack{baseViewGroup: f.baseViewGroup.clone()}
}

type FormCard struct {
	*baseViewGroup
	label          string
	supportingText string
}

func NewFormCard(id ViewID) *FormCard {
	return &FormCard{baseViewGroup: &baseViewGroup{id: id}}
}

func (f *FormCard) Label() string {
	return f.label
}

func (f *FormCard) SetLabel(s string) {
	f.label = s
}

func (f *FormCard) SupportingText() string {
	return f.supportingText
}

func (f *FormCard) SetSupportingText(s string) {
	f.supportingText = s
}

type FormCheckbox struct {
	id             ViewID
	structType     TypeID
	field          FieldID
	label          string
	supportingText string
	visible        VisibilityValueRule
}

func NewFormCheckbox(id ViewID, structType TypeID, field FieldID) *FormCheckbox {
	return &FormCheckbox{
		id:         id,
		structType: structType,
		field:      field,
	}
}

func (f *FormCheckbox) Identity() ViewID {
	return f.id
}

func (f *FormCheckbox) Clone() FormView {
	c := *f
	return &c
}

type VisibilityValueRule struct {
	Field      FieldID
	Value      string
	Comparator Comparator
}

type Comparator int

const (
	CmpEqual Comparator = iota + 1
	CmpNotEqual
	CmpLargerThan
	CmpSmallerThan
	CmpRegExp
)

func FindElementByID(root FormView, id ViewID) (FormView, bool) {
	if root.Identity() == id {
		return root, true
	}

	if root, ok := root.(FormViewGroup); ok {
		for element := range root.All() {
			if element.Identity() == id {
				return element, true
			}

			if v, ok := FindElementByID(element, id); ok {
				return v, true
			}
		}
	}

	return nil, false
}

func DeleteElementByID(root FormView, id ViewID) (FormView, bool) {
	if root, ok := root.(FormViewGroup); ok {
		for element := range root.All() {
			if element.Identity() == id {
				root.Remove(id)
				return element, true
			}

			if v, ok := DeleteElementByID(element, id); ok {
				return v, true
			}
		}
	}

	return nil, false
}

func GetViewGroup(ws *Workspace, formID FormID, vg ViewID) (FormViewGroup, bool) {
	form, ok := ws.Forms.ByID(formID)
	if !ok {
		return nil, false
	}

	parent, ok := FindElementByID(form.Root, vg)
	if !ok {
		return nil, false
	}

	parentGroup, ok := parent.(FormViewGroup)
	if !ok {
		return nil, false
	}

	return parentGroup, true
}
