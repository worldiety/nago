// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"io"
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Files = (*mistralFiles)(nil)

type mistralFiles struct {
	parent *mistralProvider
}

func (p *mistralFiles) All(subject auth.Subject) iter.Seq2[file.File, error] {
	return func(yield func(file.File, error) bool) {
		resp, err := p.client().ListFiles()
		if err != nil {
			yield(file.File{}, err)
			return
		}

		for _, f := range resp.Data {
			if !yield(f.IntoFile(), nil) {
				return
			}
		}
	}

}

func (p *mistralFiles) FindByID(subject auth.Subject, id file.ID) (option.Opt[file.File], error) {
	f, err := p.client().GetFile(string(id))
	if err != nil {
		return option.Opt[file.File]{}, err
	}

	return option.Some(f.IntoFile()), nil
}

func (p *mistralFiles) Delete(subject auth.Subject, id file.ID) error {
	return p.client().DeleteFile(string(id))
}

func (p *mistralFiles) Put(subject auth.Subject, opts file.CreateOptions) (file.File, error) {
	reader, err := opts.Open()
	if err != nil {
		return file.File{}, err
	}

	defer reader.Close()

	f, err := p.client().UploadFile(opts.Name, reader)
	if err != nil {
		return file.File{}, err
	}

	return f.IntoFile(), nil
}

func (p *mistralFiles) Get(subject auth.Subject, id file.ID) (option.Opt[io.ReadCloser], error) {
	r, err := p.client().DownloadFile(string(id))
	if err != nil {
		return option.Opt[io.ReadCloser]{}, err
	}

	return option.Some(r), nil
}

func (p *mistralFiles) client() *Client {
	return p.parent.client()
}
