// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package provider

import (
	"io"
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/auth"
)

var (
	// TooManyRequests tells you that the rate limiter has kicked in. It is the same value as
	// [completion.TooManyRequests], so the agentic helpers of the completion package can recognize it.
	TooManyRequests = completion.TooManyRequests
)

type ID string

// Provider is the central abstraction around various ai implementations like Anthropic or a local model.
type Provider interface {
	// Identity of this provider, usually based on the used secret ID.
	Identity() ID

	// Name is usually the name of the used secret.
	Name() string

	Description() string

	// Files interface to work with submitting files into the provider and reading generated files back.
	Files() option.Opt[Files]

	// Completions returns the stateless message capability, if the provider supports it. A completion always
	// receives the full history and returns a single assistant turn, like the Anthropic Messages API. The
	// available models are listed by [completion.Completions.Models].
	Completions() option.Opt[completion.Completions]
}

type Files interface {
	All(subject auth.Subject) iter.Seq2[file.File, error]
	FindByID(subject auth.Subject, id file.ID) (option.Opt[file.File], error)
	Delete(subject auth.Subject, id file.ID) error
	Put(subject auth.Subject, opts file.CreateOptions) (file.File, error)
	Get(subject auth.Subject, id file.ID) (option.Opt[io.ReadCloser], error)
}
