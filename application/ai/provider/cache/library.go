// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cache

import (
	"fmt"
	"iter"
	"os"

	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

var _ provider.Library = (*cacheLibrary)(nil)

type cacheLibrary struct {
	parent *Provider
	id     library.ID
}

func (c cacheLibrary) Identity() library.ID {
	return c.id
}

func (c cacheLibrary) All(subject auth.Subject) iter.Seq2[document.Document, error] {
	return func(yield func(document.Document, error) bool) {
		for doc, err := range c.parent.repoDocuments.All() {
			if err != nil {
				if !yield(doc, err) {
					return
				}

				continue
			}

			if doc.Library != c.id {
				continue
			}

			if doc.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoModels.Name(), string(doc.ID), PermDocumentFindAll) {
				continue
			}

			if !yield(doc, nil) {
				return
			}
		}
	}
}

func (c cacheLibrary) Delete(subject auth.Subject, id document.ID) error {
	optDoc, err := c.parent.repoDocuments.FindByID(id)
	if err != nil {
		return err
	}

	if optDoc.IsNone() {
		return nil
	}
	
	lib := optDoc.Unwrap()
	if lib.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoDocuments.Name(), string(lib.ID), PermDocumentDelete) {
		return subject.Audit(PermDocumentDelete)
	}

	if err := c.parent.prov.Libraries().Unwrap().Library(c.id).Delete(subject, id); err != nil {
		return err
	}

	return c.parent.repoDocuments.DeleteByID(id)
}

func (c cacheLibrary) Create(subject auth.Subject, opts document.CreateOptions) (document.Document, error) {
	optLib, err := c.parent.repoLibraries.FindByID(c.id)
	if err != nil {
		return document.Document{}, err
	}

	if optLib.IsNone() {
		return document.Document{}, fmt.Errorf("no such library: %s: %w", c.id, os.ErrNotExist)
	}

	lib := optLib.Unwrap()

	if subject.ID() != lib.CreatedBy && !subject.HasResourcePermission(c.parent.repoDocuments.Name(), string(lib.ID), PermDocumentCreate) {
		return document.Document{}, subject.Audit(PermDocumentCreate)
	}

	doc, err := c.parent.prov.Libraries().Unwrap().Library(c.id).Create(subject, opts)
	if err != nil {
		return document.Document{}, err
	}

	if doc.CreatedAt == 0 {
		doc.CreatedAt = xtime.Now()
	}

	doc.CreatedBy = subject.ID()
	if doc.Identity() == "" {
		return document.Document{}, fmt.Errorf("provider returned empty identity")
	}

	if opt, err := c.parent.repoDocuments.FindByID(doc.ID); err != nil || opt.IsSome() {
		if err != nil {
			return document.Document{}, err
		}

		return document.Document{}, fmt.Errorf("provider returned an existing document: %s", doc.ID)
	}

	if err := c.parent.repoDocuments.Save(doc); err != nil {
		return document.Document{}, err
	}

	return doc, nil
}
