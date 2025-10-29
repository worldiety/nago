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

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

var _ provider.Libraries = (*cacheLibraries)(nil)

type cacheLibraries struct {
	parent *Provider
}

func (c cacheLibraries) Create(subject auth.Subject, opts library.CreateOptions) (library.Library, error) {
	if err := subject.Audit(PermLibraryCreate); err != nil {
		return library.Library{}, err
	}

	lib, err := c.parent.prov.Libraries().Unwrap().Create(subject, opts)
	if err != nil {
		return library.Library{}, err
	}

	if lib.CreatedAt == 0 {
		lib.CreatedAt = xtime.Now()
	}

	lib.CreatedBy = subject.ID()
	if lib.Identity() == "" {
		return library.Library{}, fmt.Errorf("provider returned empty identity")
	}

	if opt, err := c.parent.repoLibraries.FindByID(lib.ID); err != nil || opt.IsSome() {
		if err != nil {
			return library.Library{}, err
		}

		return library.Library{}, fmt.Errorf("provider returned an existing library: %s", lib.ID)
	}

	if err := c.parent.repoLibraries.Save(lib); err != nil {
		return library.Library{}, err
	}

	return lib, nil
}

func (c cacheLibraries) FindByID(subject auth.Subject, id library.ID) (option.Opt[library.Library], error) {
	optLib, err := c.parent.repoLibraries.FindByID(id)
	if err != nil {
		return option.Opt[library.Library]{}, err
	}

	if optLib.IsNone() {
		return optLib, nil
	}

	lib := optLib.Unwrap()
	if lib.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoModels.Name(), string(lib.ID), PermLibraryFindByID) {
		return option.Opt[library.Library]{}, subject.Audit(PermLibraryFindByID)
	}

	return optLib, nil
}

func (c cacheLibraries) All(subject auth.Subject) iter.Seq2[library.Library, error] {
	return func(yield func(library.Library, error) bool) {
		for m, err := range c.parent.repoLibraries.All() {
			if err != nil {
				if !yield(m, err) {
					return
				}

				continue
			}

			if m.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoModels.Name(), string(m.ID), PermLibraryFindAll) {
				continue
			}

			if !yield(m, nil) {
				return
			}
		}
	}
}

func (c cacheLibraries) Delete(subject auth.Subject, id library.ID) error {
	optLib, err := c.parent.repoLibraries.FindByID(id)
	if err != nil {
		return err
	}

	if optLib.IsNone() {
		return nil
	}

	lib := optLib.Unwrap()
	if lib.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoModels.Name(), string(lib.ID), PermLibraryDelete) {
		return subject.Audit(PermLibraryDelete)
	}

	if err := c.parent.prov.Libraries().Unwrap().Delete(subject, id); err != nil {
		return err
	}
	
	return c.parent.repoLibraries.DeleteByID(id)
}

func (c cacheLibraries) Update(subject auth.Subject, id library.ID, opts library.UpdateOptions) (library.Library, error) {
	optLib, err := c.parent.repoLibraries.FindByID(id)
	if err != nil {
		return library.Library{}, err
	}

	if optLib.IsNone() {
		return library.Library{}, fmt.Errorf("provider returned an existing library: %s: %w", id, os.ErrNotExist)
	}

	lib := optLib.Unwrap()
	if lib.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoModels.Name(), string(lib.ID), PermLibraryUpdate) {
		return library.Library{}, subject.Audit(PermLibraryUpdate)
	}

	nLib, err := c.parent.prov.Libraries().Unwrap().Update(subject, id, opts)
	if err != nil {
		return library.Library{}, err
	}

	nLib.CreatedBy = lib.CreatedBy
	return nLib, c.parent.repoLibraries.Save(lib)
}

func (c cacheLibraries) Library(id library.ID) provider.Library {
	return &cacheLibrary{
		c.parent,
		id,
	}
}
