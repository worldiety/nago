// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"fmt"
	"iter"
	"log/slog"

	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Library = (*mistralLibrary)(nil)

type mistralLibrary struct {
	id     library.ID
	parent *mistralProvider
}

func (p *mistralLibrary) client() *Client {
	return p.parent.client()
}

func (p *mistralLibrary) Identity() library.ID {
	return p.id
}

func (p *mistralLibrary) Create(subject auth.Subject, opts document.CreateOptions) (document.Document, error) {
	doc, err := p.client().CreateDocument(string(p.id), opts.Filename, opts.Reader)

	if err != nil {
		return document.Document{}, err
	}

	slog.Info("uploaded mistral document", "id", doc.Id)

	return doc.IntoDocument(), nil
}

func (p *mistralLibrary) Delete(subject auth.Subject, doc document.ID) error {
	return p.client().DeleteDocument(string(p.id), string(doc))
}

func (p *mistralLibrary) All(subject auth.Subject) iter.Seq2[document.Document, error] {
	return func(yield func(document.Document, error) bool) {
		docs, err := p.client().ListDocuments(string(p.id))
		if err != nil {
			yield(document.Document{}, fmt.Errorf("failed to list documents from library %s: %w", p.id, err))
			return
		}

		for _, doc := range docs {
			if !yield(doc.IntoDocument(), nil) {
				return
			}
		}
	}
}
