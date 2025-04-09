// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"go.wdy.de/nago/pkg/xiter"
	"iter"
	"slices"
)

type Element interface {
	Identity() EID
	SetIdentity(id EID)
	Children() iter.Seq[Element]
	Replace(elem Element) (old Element, replaced bool)
	Append(elem Element)
	IntoPersistence() PBox
}

type EID string

type RichText struct {
	ID   EID
	Text LocStr
}

func (r *RichText) Identity() EID {
	return r.ID
}

func (r *RichText) SetIdentity(id EID) {
	r.ID = id
}

func (r *RichText) Children() iter.Seq[Element] {
	return xiter.Empty[Element]()
}

func (r *RichText) Replace(elem Element) (old Element, replaced bool) {
	return nil, false
}

func (r *RichText) Append(elem Element) {}

func (r *RichText) IntoPersistence() PBox {
	return PBox{
		RichText: &PRichText{
			ID:   r.ID,
			Text: r.Text,
		},
	}
}

type HStack struct {
	ID       EID
	Elements []Element
}

func (e *HStack) Identity() EID {
	return e.ID
}

func (e *HStack) SetIdentity(id EID) {
	e.ID = id
}

func (e *HStack) Children() iter.Seq[Element] {
	return slices.Values(e.Elements)
}

func (e *HStack) Replace(elem Element) (old Element, replaced bool) {
	for i, element := range e.Elements {
		if element.Identity() == elem.Identity() {
			e.Elements[i] = elem
			return element, true
		}
	}

	return nil, false
}

func (e *HStack) Append(elem Element) {
	e.Elements = append(e.Elements, elem)
}

func (e *HStack) IntoPersistence() PBox {
	var children []PBox
	for _, element := range e.Elements {
		children = append(children, element.IntoPersistence())
	}

	return PBox{
		HStack: &PHStack{
			ID:       e.ID,
			Children: children,
		},
	}
}

type VStack struct {
	ID       EID
	Elements []Element
}

func (e *VStack) Identity() EID {
	return e.ID
}

func (e *VStack) SetIdentity(id EID) {
	e.ID = id
}

func (e *VStack) Children() iter.Seq[Element] {
	return slices.Values(e.Elements)
}

func (e *VStack) Replace(elem Element) (old Element, replaced bool) {
	for i, element := range e.Elements {
		if element.Identity() == elem.Identity() {
			e.Elements[i] = elem
			return element, true
		}
	}

	return nil, false
}

func (e *VStack) Append(elem Element) {
	e.Elements = append(e.Elements, elem)
}

func (e *VStack) IntoPersistence() PBox {
	if e == nil {
		return PBox{}
	}
	
	var children []PBox
	for _, element := range e.Elements {
		children = append(children, element.IntoPersistence())
	}

	return PBox{
		VStack: &PVStack{
			ID:       e.ID,
			Children: children,
		},
	}
}
