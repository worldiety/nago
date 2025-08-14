// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package document

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/signature"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xslices"
	"iter"
	"time"
)

type ID string
type State int

const (
	Unknown State = iota
)

type Participant struct {
	Firstname string
	Lastname  string
	Email     user.Email
	User      option.Opt[user.ID]
}

type Document struct {
	ID           ID
	CreatedAt    time.Time
	CreatedBy    user.ID
	Name         string
	Participants xslices.Slice[Participant]
}

func (d Document) Identity() ID {
	return d.ID
}

type CreationData struct {
	Name         string
	Participants []Participant
}
type Create func(subject user.Subject, document Document) (ID, error)
type FindByMail func(subject user.Subject, mail user.Email) iter.Seq2[Document, error]
type Sign func(subject user.Subject, doc ID, signature signature.ID) (ID, error)
type Signed func(subject user.Subject, mail user.Email) (bool, error)
