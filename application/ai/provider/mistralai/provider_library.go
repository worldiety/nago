// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"fmt"
	"io"
	"iter"
	"log/slog"
	"strings"

	"github.com/worldiety/option"
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
		if strings.Contains(strings.ToLower(err.Error()), "mimetype not supported") {
			return document.Document{}, fmt.Errorf("%w: %w", err, document.UnsupportedFormatError)
		}

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

func (p *mistralLibrary) TextContentByID(subject auth.Subject, id document.ID) (option.Opt[string], error) {
	text, err := p.client().GetDocumentText(string(p.id), string(id))
	if err != nil {
		return option.Opt[string]{}, err
	}

	return option.Some(text), nil
}

func (p *mistralLibrary) StatusByID(subject auth.Subject, id document.ID) (option.Opt[document.ProcessingStatus], error) {
	status, err := p.client().GetDocumentStatus(string(p.id), string(id))
	if err != nil {
		return option.Opt[document.ProcessingStatus]{}, err
	}

	return option.Some(status), nil
}

func (p *mistralLibrary) FindByID(subject auth.Subject, id document.ID) (option.Opt[document.Document], error) {
	doc, err := p.client().GetDocument(string(p.id), string(id))
	if err != nil {
		return option.Opt[document.Document]{}, err
	}

	return option.Some(doc.IntoDocument()), nil
}

func (p *mistralLibrary) Get(subject auth.Subject, id document.ID) (option.Opt[io.ReadCloser], error) {
	reader, err := p.client().GetDocumentDownload(string(p.id), string(id))
	if err != nil {
		return option.Opt[io.ReadCloser]{}, err
	}

	return option.Some(reader), nil
}
