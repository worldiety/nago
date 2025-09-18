// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package video

import (
	"io"
	"iter"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/video/channel"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

type ID string
type Video struct {
	ID          ID            `json:"id,omitempty"`
	Title       string        `json:"title,omitempty"`
	Description string        `json:"description,omitempty"`
	Thumbnail   image.ID      `json:"thumbnail,omitempty"`
	Length      time.Duration `json:"length,omitempty"`
	Public      bool          `json:"public,omitempty"`
	Sources     []Source      `json:"sources,omitempty"`
}

func (v Video) Identity() ID {
	return v.ID
}

// Source represents an encoded variant of a video.
// See also https://www.w3.org/TR/2011/WD-html5-20110113/video.html#the-source-element
type Source struct {
	Blob     string   // id to the blob store
	Filename string   // e.g. video.mp4
	Sha3H256 string   // Sha3H256
	Width    int      // in pixel
	Height   int      // in pixel
	Size     int64    // in bytes
	MimeType string   // e.g. video/mp4
	Codecs   []string // e.g. avc1.4D401E, mp4a.40.2
}

type CreateOptions struct {
	Title       string
	Description string
	Public      bool
}

type Create func(subject auth.Subject, opts CreateOptions) (ID, error)
type Delete func(subject auth.Subject, id ID) error
type FindByID func(subject auth.Subject, id ID) (option.Opt[Video], error)
type FindAll func(subject auth.Subject) iter.Seq2[Video, error]

type AddSource func(subject auth.Subject, id ID, name string, reader io.Reader) error
type DeleteSource func(subject auth.Subject, id ID) error

type AddToChannel func(subject auth.Subject, id channel.ID, video ID) error
type RemoveFromChannel func(subject auth.Subject, id channel.ID, video ID) error

type Repository data.Repository[Video, ID]

type UseCases struct {
	Create Create
}

func NewUseCases(repo Repository) UseCases {
	return UseCases{}
}
