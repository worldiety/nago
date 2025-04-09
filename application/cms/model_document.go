// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"iter"
	"time"
)

// A Document represents an unstructured content tree which at least provides a linear slice of elements which
// must be rendered in the defined order. We do not model permissions here, because we rely on the resource-based
// permission system.
type Document struct {
	ID          ID
	Slug        Slug
	LastUpdated time.Time
	Title       LocStr
	Body        *VStack
	Published   bool
}

func (p *Document) String() string {
	return p.Title.String()
}

func (p *Document) IntoPersistence() PDocument {
	return PDocument{
		ID:          p.ID,
		Slug:        p.Slug,
		LastUpdated: p.LastUpdated,
		Title:       p.Title,
		Published:   p.Published,
		Body:        p.Body.IntoPersistence().VStack,
	}
}

func (p *Document) ElementByID(id EID) (Element, bool) {
	for element := range Visit(p.Body) {
		if element.Identity() == id {
			return element, true
		}
	}

	return nil, false
}

func (p *Document) ParentOf(id EID) (Element, bool) {
	for element := range Visit(p.Body) {
		for c := range element.Children() {
			if c.Identity() == id {
				return element, true
			}
		}
	}

	return nil, false
}

func (p *Document) Replace(elem Element) bool {
	root, ok := p.ParentOf(elem.Identity())
	if !ok {
		return false
	}

	_, ok = root.Replace(elem)
	return ok
}

func (p *Document) Append(parent EID, elem Element) bool {
	root, ok := p.ParentOf(parent)
	if !ok {
		return false
	}

	root.Append(elem)
	return true
}

func (p *Document) Identity() ID {
	if p == nil {
		return ""
	}
	return p.ID
}

func Visit(root Element) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		visitRec(root, yield)
	}
}

func visitRec(root Element, visit func(Element) bool) bool {
	if !visit(root) {
		return false
	}

	for element := range root.Children() {
		if !visitRec(element, visit) {
			return false
		}
	}

	return true
}
