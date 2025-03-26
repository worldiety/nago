// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package docm

type Element interface {
	isElement()
}

type Heading struct {
	Level int
	Body  Element
	Label Label
}

func (e *Heading) isElement() {
}

type Pre struct {
	Lines []string
}

func (e *Pre) Add(s string) {
	e.Lines = append(e.Lines, s)
}

func (e *Pre) isElement() {
}

type Text struct {
	Value string
}

func (e *Text) isElement() {
}

// Sequence is just bunch of Element without any special rules.
type Sequence []Element

func (e Sequence) Last() (Element, bool) {
	if len(e) == 0 {
		return nil, false
	}

	return e[len(e)-1], true
}

func (e *Sequence) Add(elem Element) {
	*e = append(*e, elem)
}

func (e Sequence) isElement() {
}

// A List is used for enumerations like a bullet or ordered sequence of elements.
type List struct {
	Children []Element
}

func (e *List) isElement() {
}

func (e *List) Add(element Element) {
	e.Children = append(e.Children, element)
}

type Par struct {
	Children []Element
}

func (e *Par) Add(element Element) {
	e.Children = append(e.Children, element)
}

func (e *Par) isElement() {
}

type Link struct {
	Body Element
	Dest Label
}

func (e *Link) isElement() {
}

type Label string

// Document is the root for all other Elements.
type Document struct {
	Body Element
}

type Image struct {
	URL string
}

func (e *Image) isElement() {}
