// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cache

import (
	"fmt"
	"io"
	"iter"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
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
		var docs []document.Document
		// collect our visible docs
		for doc, err := range c.parent.repoDocuments.All() {
			if err != nil {
				yield(doc, err)
				return
			}

			if doc.Library != c.id {
				continue
			}

			if doc.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoModels.Name(), string(doc.ID), PermDocumentFindAll) {
				continue
			}

			docs = append(docs, doc)
		}

		// check if we need a reload
		needsReload := map[document.ID]document.Document{}
		for _, doc := range docs {
			if doc.ProcessingStatus == document.ProcessingRunning {
				needsReload[doc.ID] = doc
			}
		}

		// update required ones
		if len(needsReload) > 0 {
			for id, doc := range needsReload {
				optStat, err := c.parent.prov.Libraries().Unwrap().Library(c.id).StatusByID(subject, id)
				if err != nil {
					yield(document.Document{}, err)
					return
				}

				if optStat.IsNone() {
					if err := c.parent.repoDocuments.DeleteByID(id); err != nil {
						yield(document.Document{}, err)
						return
					}

					continue
				}

				stat := optStat.Unwrap()
				if stat != doc.ProcessingStatus {
					doc.ProcessingStatus = stat
					if err := c.parent.repoDocuments.Save(doc); err != nil {
						yield(document.Document{}, err)
						return
					}
				}
			}

			// collect again
			docs = docs[:0]
			for doc, err := range c.parent.repoDocuments.All() {
				if err != nil {
					yield(doc, err)
					return
				}

				if doc.Library != c.id {
					continue
				}

				if doc.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoModels.Name(), string(doc.ID), PermDocumentFindAll) {
					continue
				}

				docs = append(docs, doc)
			}
		}

		slices.SortFunc(docs, func(a, b document.Document) int {
			return strings.Compare(a.Name, b.Name)
		})

		for _, doc := range docs {
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

		slog.Warn("provider returned an existing document, this may be intentional (e.g. if identical document was uploaded) or an unwanted collision", "doc", doc.ID)
		//return document.Document{}, fmt.Errorf("provider returned an existing document: %s", doc.ID)
	}

	if err := c.parent.repoDocuments.Save(doc); err != nil {
		return document.Document{}, err
	}

	return doc, nil
}

func (c cacheLibrary) TextContentByID(subject auth.Subject, id document.ID) (option.Opt[string], error) {
	optDoc, err := c.parent.repoDocuments.FindByID(id)
	if err != nil {
		return option.Opt[string]{}, err
	}

	if optDoc.IsNone() {
		return option.Opt[string]{}, nil
	}

	doc := optDoc.Unwrap()
	if doc.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoDocuments.Name(), string(doc.ID), PermReadTextContent) {
		return option.Opt[string]{}, subject.Audit(PermReadTextContent)
	}

	optText, err := blob.Get(c.parent.docTextStore, string(id))
	if err != nil {
		return option.Opt[string]{}, err
	}

	if optText.IsSome() {
		return option.Some(string(optText.Unwrap())), nil
	}

	optStrText, err := c.parent.prov.Libraries().Unwrap().Library(c.id).TextContentByID(subject, id)
	if err != nil {
		return option.Opt[string]{}, err
	}

	if optStrText.IsNone() {
		return option.Opt[string]{}, nil
	}

	if err := blob.Put(c.parent.docTextStore, string(id), []byte(optStrText.Unwrap())); err != nil {
		return option.Opt[string]{}, err
	}

	return optStrText, nil
}

func (c cacheLibrary) StatusByID(subject auth.Subject, id document.ID) (option.Opt[document.ProcessingStatus], error) {
	optDoc, err := c.parent.repoDocuments.FindByID(id)
	if err != nil {
		return option.Opt[document.ProcessingStatus]{}, nil
	}

	if optDoc.IsNone() {
		return option.Opt[document.ProcessingStatus]{}, nil
	}

	doc := optDoc.Unwrap()
	if doc.ProcessingStatus != document.ProcessingCompleted {
		optStat, err := c.parent.prov.Libraries().Unwrap().Library(c.id).StatusByID(subject, id)
		if err != nil {
			return option.Opt[document.ProcessingStatus]{}, err
		}

		if optStat.IsNone() {
			if err := c.parent.repoDocuments.DeleteByID(id); err != nil {
				return option.Opt[document.ProcessingStatus]{}, err
			}

			return option.Opt[document.ProcessingStatus]{}, nil
		}

		stat := optStat.Unwrap()
		if stat != doc.ProcessingStatus {
			doc.ProcessingStatus = stat
			if err := c.parent.repoDocuments.Save(doc); err != nil {
				return option.Opt[document.ProcessingStatus]{}, err
			}
		}
	}

	return option.Some(doc.ProcessingStatus), nil
}

func (c cacheLibrary) FindByID(subject auth.Subject, id document.ID) (option.Opt[document.Document], error) {
	optDoc, err := c.parent.repoDocuments.FindByID(id)
	if err != nil {
		return option.Opt[document.Document]{}, err
	}

	if optDoc.IsNone() {
		return option.Opt[document.Document]{}, nil
	}

	doc := optDoc.Unwrap()
	if doc.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoDocuments.Name(), string(doc.ID), PermDocumentFindAll) {
		return option.Opt[document.Document]{}, subject.Audit(PermDocumentFindAll)
	}

	return option.Some(doc), nil
}

func (c cacheLibrary) Get(subject auth.Subject, id document.ID) (option.Opt[io.ReadCloser], error) {
	// TODO should we store the file again? If source is our drive we would need to double the files and occupied space
	return c.parent.prov.Libraries().Unwrap().Library(c.id).Get(subject, id)
}
