// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package channel

import (
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

type ID string

type Channel struct {
	ID          ID       `json:"id"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Image       image.ID `json:"image,omitempty"`
}

func (c Channel) Identity() ID {
	return c.ID
}

type CreateOptions struct {
	Title       string
	Description string
	Image       image.ID
}

type Create func(subject auth.Subject, options CreateOptions) (ID, error)
type Delete func(subject auth.Subject, id ID) error

type FindAll func(subject auth.Subject) iter.Seq2[Channel, error]

type FindByID func(subject auth.Subject, id ID) (option.Opt[Channel], error)

type Repository data.Repository[Channel, ID]
