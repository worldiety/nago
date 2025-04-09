// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"time"
)

type Repository = data.Repository[PDocument, ID]
type PDocument struct {
	ID          ID        `json:"id,omitempty"`
	Slug        Slug      `json:"slug,omitempty"`
	Title       LocStr    `json:"title,omitempty"`
	LastUpdated time.Time `json:"lastUpdated,omitempty"`
	Published   bool      `json:"published,omitempty"`
	Body        *PVStack  `json:"body,omitempty"`
}

func (p PDocument) IntoModel() *Document {
	doc := &Document{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		LastUpdated: p.LastUpdated,
		Published:   p.Published,
		Body:        p.Body.IntoModel(),
	}

	return doc
}

func (p PDocument) Identity() ID {
	return p.ID
}

type PBox struct {
	VStack   *PVStack   `json:"vstack,omitempty"`
	HStack   *PHStack   `json:"hstack,omitempty"`
	RichText *PRichText `json:"richText,omitempty"`
}

func (p PBox) IntoModel() Element {
	switch {
	case p.VStack != nil:
		return p.VStack.IntoModel()
	case p.HStack != nil:
		return p.HStack.IntoModel()
	case p.RichText != nil:
		return p.RichText.IntoModel()
	default:
		panic(fmt.Errorf("unknown model type: %T", p))
	}
}

type PVStack struct {
	ID       EID    `json:"id"`
	Children []PBox `json:"children"`
}

func (p *PVStack) IntoModel() *VStack {
	v := &VStack{
		ID: p.ID,
	}

	for _, child := range p.Children {
		v.Elements = append(v.Elements, child.IntoModel())
	}

	return v
}

type PHStack struct {
	ID       EID    `json:"id"`
	Children []PBox `json:"children"`
}

func (p *PHStack) IntoModel() *VStack {
	v := &VStack{
		ID: p.ID,
	}

	for _, child := range p.Children {
		v.Elements = append(v.Elements, child.IntoModel())
	}

	return v
}

type PRichText struct {
	ID   EID    `json:"id"`
	Text LocStr `json:"text"`
}

func (p *PRichText) IntoModel() *RichText {
	return &RichText{
		ID:   p.ID,
		Text: p.Text,
	}
}
