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

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Libraries = (*mistralLibraries)(nil)

type mistralLibraries struct {
	parent *mistralProvider
}

func (p *mistralLibraries) Library(id library.ID) provider.Library {
	return &mistralLibrary{
		id:     id,
		parent: p.parent,
	}
}

func (p *mistralLibraries) Create(subject auth.Subject, opts library.CreateOptions) (library.Library, error) {
	info, err := p.client().CreateLibrary(CreateLibraryRequest{
		Description: opts.Description,
		Name:        opts.Name,
	})

	if err != nil {
		return library.Library{}, err
	}

	if info.Id == "" {
		return library.Library{}, fmt.Errorf("failed to create library: received empty id, probably a mistral protocol error")
	}

	return info.IntoLibrary(), nil
}

func (p *mistralLibraries) Update(subject auth.Subject, id library.ID, opts library.UpdateOptions) (library.Library, error) {
	info, err := p.client().UpdateLibrary(string(id), UpdateLibraryRequest{
		Description: &opts.Description,
		Name:        &opts.Name,
	})

	if err != nil {
		return library.Library{}, err
	}

	if info.Id == "" {
		return library.Library{}, fmt.Errorf("failed to create library: received empty id, probably a mistral protocol error")
	}

	return info.IntoLibrary(), nil
}

func (p *mistralLibraries) FindByID(subject auth.Subject, id library.ID) (option.Opt[library.Library], error) {
	info, err := p.client().GetLibrary(string(id))
	if err != nil {
		return option.Opt[library.Library]{}, err
	}

	if info.Id == "" {
		return option.Opt[library.Library]{}, fmt.Errorf("failed to get library: received empty id, probably a mistral protocol error")
	}

	return option.Some(info.IntoLibrary()), nil
}

func (p *mistralLibraries) All(subject auth.Subject) iter.Seq2[library.Library, error] {
	return func(yield func(library.Library, error) bool) {
		infos, err := p.client().ListAllLibraries()
		if err != nil {
			yield(library.Library{}, err)
			return
		}

		for _, info := range infos {
			if !yield(info.IntoLibrary(), nil) {
				return
			}
		}
	}
}

func (p *mistralLibraries) Delete(subject auth.Subject, id library.ID) error {
	return p.client().DeleteLibrary(string(id))
}

func (p *mistralLibraries) client() *Client {
	return p.parent.client()
}
