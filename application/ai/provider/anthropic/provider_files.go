// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"io"
	"iter"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

var _ provider.Files = (*anthropicFiles)(nil)

// anthropicFiles implements [provider.Files] on top of the Anthropic Files API (beta). It supports uploading,
// listing, inspecting and deleting files. Files uploaded here can be referenced from message content by their
// id (see [completion.FileRef]), so binary content is sent to Anthropic once and then referenced O(1) per
// turn instead of being inlined as base64.
//
// Note: Anthropic only allows downloading API/tool-generated files. Downloading a user-uploaded file yields
// [file.ErrNotDownloadable].
type anthropicFiles struct {
	parent *anthropicProvider
}

func (f *anthropicFiles) client() *Client {
	return f.parent.client()
}

func intoFile(m apiFileMetadata) file.File {
	var createdAt xtime.UnixMilliseconds
	if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
		createdAt = xtime.UnixMilliseconds(t.UnixMilli())
	}

	return file.File{
		ID:        file.ID(m.ID),
		Name:      m.FileName,
		MimeType:  file.Type(m.MimeType),
		CreatedAt: createdAt,
	}
}

func (f *anthropicFiles) All(subject auth.Subject) iter.Seq2[file.File, error] {
	return func(yield func(file.File, error) bool) {
		metas, err := f.client().ListFiles()
		if err != nil {
			yield(file.File{}, err)
			return
		}

		for _, m := range metas {
			if !yield(intoFile(m), nil) {
				return
			}
		}
	}
}

func (f *anthropicFiles) FindByID(subject auth.Subject, id file.ID) (option.Opt[file.File], error) {
	m, err := f.client().GetFileMetadata(string(id))
	if err != nil {
		return option.Opt[file.File]{}, err
	}

	return option.Some(intoFile(m)), nil
}

func (f *anthropicFiles) Delete(subject auth.Subject, id file.ID) error {
	return f.client().DeleteFile(string(id))
}

func (f *anthropicFiles) Put(subject auth.Subject, opts file.CreateOptions) (file.File, error) {
	reader, err := opts.Open()
	if err != nil {
		return file.File{}, err
	}
	defer reader.Close()

	m, err := f.client().UploadFile(opts.Name, string(opts.MimeType), reader)
	if err != nil {
		return file.File{}, err
	}

	return intoFile(m), nil
}

// Get streams the raw file bytes. Anthropic only permits this for API/tool-generated files; user-uploaded
// files respond with a 400 which is translated to [file.ErrNotDownloadable].
func (f *anthropicFiles) Get(subject auth.Subject, id file.ID) (option.Opt[io.ReadCloser], error) {
	rc, err := f.client().DownloadFile(string(id))
	if err != nil {
		return option.Opt[io.ReadCloser]{}, err
	}

	return option.Some(rc), nil
}
