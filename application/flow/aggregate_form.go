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

	"go.wdy.de/nago/presentation/ui"
)

type Form struct {
	ID             FormID
	Root           FormView
	name           Ident
	description    string
	repository     RepositoryID
	repositoryType TypeID
}

func NewForm(id FormID, name Ident, root FormView, repository RepositoryID, repositoryType TypeID) *Form {
	return &Form{
		ID:             id,
		name:           name,
		Root:           root,
		repository:     repository,
		repositoryType: repositoryType,
	}
}

func (f *Form) Repository() RepositoryID {
	return f.repository
}

func (f *Form) RepositoryType() TypeID {
	return f.repositoryType
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
	c := &Form{
		ID:             f.ID,
		name:           f.name,
		description:    f.description,
		repository:     f.repository,
		repositoryType: f.repositoryType,
	}

	if f.Root != nil {
		c.Root = f.Root.Clone()
	}
	return c
}

type Expression string

type FormView interface {
	Identity() ViewID
	Clone() FormView
	VisibleExpr() Expression
	SetVisibleExpr(Expression)
}

// FormViewGroup extens FormView with view group parent functions. A view tree must not contain cycles.
type FormViewGroup interface {
	FormView
	Len() int
	All() iter.Seq[FormView]
	// Insert merges the given view into the ordered children list.
	// If after is empty, the view is appended to the beginning of the group.
	Insert(view FormView, after ViewID)
	Remove(ViewID)
}

type Alignable interface {
	FormView
	Alignment() ui.Alignment
	SetAlignment(ui.Alignment)
}

type Actionable interface {
	FormView
	ActionExpr() []Expression
	SetActionExpr(...Expression)
}

type Enabler interface {
	FormView
	EnabledExpr() Expression
	SetEnabledExpr(Expression)
}

type Gapable interface {
	FormView
	Gap() ui.Length
	SetGap(ui.Length)
}

type ViewID string

var (
	_ FormView      = (*FormText)(nil)
	_ FormView      = (*FormCheckbox)(nil)
	_ FormViewGroup = (*baseViewGroup)(nil)
	_ FormViewGroup = (*FormCard)(nil)
)

type baseViewGroup struct {
	id          ViewID
	views       []FormView
	visibleExpr Expression
}

func (b *baseViewGroup) Identity() ViewID {
	return b.id
}

func (b *baseViewGroup) Clone() FormView {
	return b.clone()
}

func (b *baseViewGroup) clone() *baseViewGroup {
	return &baseViewGroup{
		id:          b.id,
		views:       slices.Clone(b.views),
		visibleExpr: b.visibleExpr,
	}
}

func (b *baseViewGroup) VisibleExpr() Expression {
	return b.visibleExpr
}

func (b *baseViewGroup) SetVisibleExpr(expr Expression) {
	b.visibleExpr = expr
}

func (b *baseViewGroup) All() iter.Seq[FormView] {
	return slices.Values(b.views)
}

func (b *baseViewGroup) Len() int {
	return len(b.views)
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

func GetView(ws *Workspace, formID FormID, v ViewID) (FormView, bool) {
	form, ok := ws.Forms.ByID(formID)
	if !ok {
		return nil, false
	}

	return FindElementByID(form.Root, v)
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
