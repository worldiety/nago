// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/std/concurrent"
	"golang.org/x/text/language"
	"iter"
	"regexp"
	"strings"
	"sync"
)

type ID string

// Slug represents the URL-friendly identifier for a web page.
// It is typically a lowercase string with words separated by hyphens,
// used to identify the page in a readable and SEO-friendly format.
// For example, a page titled "How to Make Pizza" might have the slug
// "how-to-make-pizza", resulting in a URL like "/blog/how-to-make-pizza".
type Slug string

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s Slug) Validate() error {
	if len(strings.TrimSpace(string(s))) == 0 {
		return std.NewLocalizedError("Ungültiger Slug", "Slug darf nicht leer sein.")
	}

	if !slugRegex.MatchString(string(s)) {
		return std.NewLocalizedError("Ungültiger Slug", "Slug hast ein ungültiges Format und muss aussehen wie abc oder ab-bc-cefg.")
	}

	return nil
}

type LocStr map[language.Tag]string

func (s LocStr) String() string {
	if s == nil {
		return ""
	}

	if v, ok := s[language.German]; ok {
		return v
	}

	if v, ok := s[language.English]; ok {
		return v
	}

	if v, ok := s[language.Und]; ok {
		return v
	}

	for _, s := range s {
		return s
	}

	return ""
}

func (s LocStr) Match(lang language.Tag) string {
	if s == nil {
		return ""
	}

	if s, ok := s[lang]; ok {
		return s
	}

	return s.String()
}

type CreationData struct {
	Title     string `label:"Seitentitel"`
	Slug      Slug   `label:"Slug"`
	Published bool   `label:"Veröffentlicht"`
}

type Create func(subject auth.Subject, data CreationData) (ID, error)
type Delete func(subject auth.Subject, id ID) error

type UpdateSlug func(subject auth.Subject, id ID, slug Slug) error

type UpdateTitle func(subject auth.Subject, id ID, lang language.Tag, title string) error

type UpdatePublished func(subject auth.Subject, id ID, published bool) error

type UpdateElement func(subject auth.Subject, id ID, element EID, mutator func(elem Element) Element) error

type ReplaceElement func(subject auth.Subject, id ID, element Element) error

type DeleteElement func(subject auth.Subject, id ID, element EID) error

type AppendElement func(subject auth.Subject, id ID, parent EID, elem Element) error

type FindAll func(subject auth.Subject) iter.Seq2[*Document, error]
type FindByID func(subject auth.Subject, id ID) (option.Opt[*Document], error)
type FindBySlug func(subject auth.Subject, slug Slug) (option.Opt[*Document], error)

type UseCases struct {
	Create          Create
	Delete          Delete
	UpdateSlug      UpdateSlug
	UpdateTitle     UpdateTitle
	UpdatePublished UpdatePublished
	UpdateElement   UpdateElement
	ReplaceElement  ReplaceElement
	DeleteElement   DeleteElement
	AppendElement   AppendElement
	FindAll         FindAll
	FindByID        FindByID
	FindBySlug      FindBySlug
}

func NewUseCases(repo Repository) (UseCases, error) {
	slugReverseLookup := &concurrent.RWMap[Slug, ID]{}
	var mutex sync.Mutex

	for doc, err := range repo.All() {
		if err != nil {
			return UseCases{}, err
		}

		if id, ok := slugReverseLookup.Get(doc.Slug); ok {
			return UseCases{}, fmt.Errorf("slug is ambigious: '%s.%s' vs '%s.%s'", doc.ID, doc.Slug, id, doc.Slug)
		}

		slugReverseLookup.Put(doc.Slug, doc.ID)
	}

	return UseCases{
		Create:          NewCreate(&mutex, slugReverseLookup, repo),
		Delete:          NewDelete(&mutex, slugReverseLookup, repo),
		UpdateSlug:      NewUpdateSlug(&mutex, slugReverseLookup, repo),
		UpdateTitle:     NewUpdateTitle(&mutex, repo),
		UpdatePublished: NewUpdatePublished(&mutex, repo),
		UpdateElement:   NewUpdateElement(&mutex, repo),
		FindAll:         NewFindAll(repo),
		AppendElement:   NewAppendElement(&mutex, repo),
		FindByID:        NewFindByID(repo),
		FindBySlug:      NewFindBySlug(slugReverseLookup, repo),
		ReplaceElement:  NewReplaceElement(&mutex, repo),
	}, nil
}
