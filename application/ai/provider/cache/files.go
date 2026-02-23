// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cache

import (
	"context"
	"fmt"
	"io"
	"iter"
	"log/slog"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/xtime"
)

var _ provider.Files = (*cacheFiles)(nil)

type cacheFiles struct {
	parent *Provider
}

func (c cacheFiles) All(subject auth.Subject) iter.Seq2[file.File, error] {
	return func(yield func(file.File, error) bool) {
		for key, err := range c.parent.idxProvFiles.AllByPrimary(context.Background(), c.parent.Identity()) {
			if err != nil {
				if !yield(file.File{}, err) {
					return
				}

				continue
			}

			optConv, err := c.parent.repoFiles.FindByID(key.Secondary)
			if err != nil {
				if !yield(file.File{}, err) {
					return
				}

				continue
			}

			if optConv.IsNone() {
				continue // stale ref
			}

			m := optConv.Unwrap()

			if m.CreatedBy != subject.ID() && !subject.HasResourcePermission(rebac.Namespace(c.parent.repoFiles.Name()), rebac.Instance(m.ID), PermFileFindAll) {
				continue
			}

			if !yield(m, nil) {
				return
			}
		}
	}
}

func (c cacheFiles) FindByID(subject auth.Subject, id file.ID) (option.Opt[file.File], error) {
	optLib, err := c.parent.repoFiles.FindByID(id)
	if err != nil {
		return option.Opt[file.File]{}, err
	}

	if optLib.IsNone() {
		return optLib, nil
	}

	lib := optLib.Unwrap()
	if lib.CreatedBy != subject.ID() && !subject.HasResourcePermission(rebac.Namespace(c.parent.repoFiles.Name()), rebac.Instance(lib.ID), PermFileFindByID) {
		return option.Opt[file.File]{}, subject.Audit(PermFileFindByID)
	}

	return optLib, nil
}

func (c cacheFiles) Delete(subject auth.Subject, id file.ID) error {
	optLib, err := c.parent.repoFiles.FindByID(id)
	if err != nil {
		return err
	}

	if optLib.IsNone() {
		return nil
	}

	lib := optLib.Unwrap()
	if lib.CreatedBy != subject.ID() && !subject.HasResourcePermission(rebac.Namespace(c.parent.repoFiles.Name()), rebac.Instance(lib.ID), PermFileDelete) {
		return subject.Audit(PermFileDelete)
	}

	if err := c.parent.prov.Files().Unwrap().Delete(subject, id); err != nil {
		return err
	}

	if err := c.parent.idxProvFiles.Delete(context.Background(), c.parent.Identity(), lib.ID); err != nil {
		return err
	}

	if err := c.parent.fileStore.Delete(context.Background(), string(lib.ID)); err != nil {
		return err
	}

	return c.parent.repoFiles.DeleteByID(id)
}

func (c cacheFiles) Put(subject auth.Subject, opts file.CreateOptions) (file.File, error) {
	if err := subject.Audit(PermFilePut); err != nil {
		return file.File{}, err
	}

	doc, err := c.parent.prov.Files().Unwrap().Put(subject, opts)
	if err != nil {
		return file.File{}, err
	}

	if doc.CreatedAt == 0 {
		doc.CreatedAt = xtime.Now()
	}

	doc.CreatedBy = subject.ID()
	if doc.Identity() == "" {
		return file.File{}, fmt.Errorf("provider returned empty identity")
	}

	if opt, err := c.parent.repoFiles.FindByID(doc.ID); err != nil || opt.IsSome() {
		if err != nil {
			return file.File{}, err
		}

		slog.Warn("provider returned an existing file, this may be intentional (e.g. if identical file was uploaded) or an unwanted collision", "file", doc.ID)
	}

	if err := c.parent.repoFiles.Save(doc); err != nil {
		return file.File{}, err
	}

	reader, err := opts.Open()
	if err != nil {
		return file.File{}, err
	}

	defer reader.Close()

	if _, err := blob.Write(c.parent.fileStore, string(doc.ID), reader); err != nil {
		return file.File{}, err
	}

	if err := c.parent.idxProvFiles.Put(c.parent.Identity(), doc.ID); err != nil {
		return file.File{}, err
	}

	return doc, nil
}

func (c cacheFiles) Get(subject auth.Subject, id file.ID) (option.Opt[io.ReadCloser], error) {
	if !subject.HasPermission(PermFileGet) {
		// do we have file ownership?
		optFile, err := c.FindByID(subject, id)
		if err != nil {
			return option.None[io.ReadCloser](), err
		}

		if optFile.IsNone() {
			return option.None[io.ReadCloser](), nil
		}
	}

	// try to find in cache anyway
	optReader, err := c.parent.fileStore.NewReader(context.Background(), string(id))
	if err != nil {
		return option.None[io.ReadCloser](), err
	}

	if optReader.IsSome() {
		return optReader, nil
	}

	// not in cache, but may have been created by the provider in its cloud space
	optFile, err := c.parent.prov.Files().Unwrap().FindByID(subject, id)
	if err != nil {
		return option.None[io.ReadCloser](), err
	}

	if optFile.IsNone() {
		return option.None[io.ReadCloser](), nil
	}

	f := optFile.Unwrap()

	optReader, err = c.parent.prov.Files().Unwrap().Get(subject, id)
	if err != nil {
		return option.None[io.ReadCloser](), err
	}

	if optReader.IsNone() {
		return option.None[io.ReadCloser](), nil
	}

	reader := optReader.Unwrap()
	defer reader.Close()

	// insert into our cache
	if _, err := blob.Write(c.parent.fileStore, string(id), reader); err != nil {
		return option.None[io.ReadCloser](), err
	}

	if err := c.parent.repoFiles.Save(f); err != nil {
		return option.None[io.ReadCloser](), err
	}

	if err := c.parent.idxProvFiles.Put(c.parent.Identity(), f.Identity()); err != nil {
		return option.None[io.ReadCloser](), err
	}

	// re-read from cache
	return c.parent.fileStore.NewReader(context.Background(), string(id))
}
