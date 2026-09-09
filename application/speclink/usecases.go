// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package speclink

import "io/fs"

// UseCases bundles the capabilities of the speclink context.
//
// There is nothing to thread through: the catalogue and the binding registry are populated by package
// initialisation of whatever the binary links in, so the use cases read a global that Go itself filled.
// That is unusual for nago and it is the honest shape here - a repository would be a pretence of
// configurability that does not exist.
type UseCases struct {
	FindAllRequirements FindAllRequirements
	FindRequirementByID FindRequirementByID
	FindAllCapabilities FindAllCapabilities
	FindSourceDocument  FindSourceDocument
}

// NewUseCases wires the speclink use cases.
//
// sources is the filesystem holding the documents the requirements were derived from, usually an embed.FS
// over requirements/_sources. It may be nil; the reading use case then reports that none were embedded.
func NewUseCases(sources fs.FS) UseCases {
	return UseCases{
		FindAllRequirements: NewFindAllRequirements(),
		FindRequirementByID: NewFindRequirementByID(),
		FindAllCapabilities: NewFindAllCapabilities(),
		FindSourceDocument:  NewFindSourceDocument(sources),
	}
}
