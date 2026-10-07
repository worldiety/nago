// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

// ProviderFileUploader adapts a provider's Files capability into a [completion.FileUploader] so file-providing
// tools (see [completion.NewOpenFileTool]) can upload a file to the provider and reference it by id. It
// returns nil when the provider exposes no Files capability, which makes such tool calls report a friendly
// error to the model instead of attaching anything.
func ProviderFileUploader(prov provider.Provider) completion.FileUploader {
	return ProviderFileUploaderFor(prov, "")
}

// ProviderFileUploaderFor is like [ProviderFileUploader], but uploads the files on behalf of the owner, usually
// the session of the run, so that they are deleted at the provider together with the session, see
// [file.CreateOptions.Owner].
func ProviderFileUploaderFor(prov provider.Provider, owner file.Owner) completion.FileUploader {
	optFiles := prov.Files()
	if optFiles.IsNone() {
		return nil
	}
	files := optFiles.Unwrap()
	return func(subject auth.Subject, f completion.OpenedFile) (file.ID, error) {
		uploaded, err := files.Put(subject, file.CreateOptions{
			Name:     f.Name,
			MimeType: f.MimeType,
			Purpose:  file.PurposeUserData,
			Open:     f.Open,
			Owner:    owner,
		})
		if err != nil {
			return "", err
		}
		return uploaded.ID, nil
	}
}
