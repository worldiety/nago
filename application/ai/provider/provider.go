// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package provider

import (
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/auth"
)

type ID string

type Provider interface {
	// Identity of this provider, usually based on the used secret ID.
	Identity() ID

	// Name is usually the name of the used secret.
	Name() string

	Description() string

	// Libraries returns the implementation, if this Provider supports native libraries.
	Libraries() option.Opt[Libraries]

	Agents() option.Opt[Agents]
}

type Libraries interface {
	Create(subject auth.Subject, opts library.CreateOptions) (library.Library, error)
	FindByID(subject auth.Subject, id library.ID) (option.Opt[library.Library], error)
	All(subject auth.Subject) iter.Seq2[library.Library, error]
	Delete(subject auth.Subject, id library.ID) error
}

type Agents interface {
	All(subject auth.Subject) iter.Seq2[agent.Agent, error]
	Delete(subject auth.Subject, id agent.ID) error
}
