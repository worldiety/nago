// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package localization

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"golang.org/x/text/language"
)

// Path is a part of a [i18n.Key]. Path segments are separated by dot (.) instead of /. The root path is just
// the dot. The basis for this kind of hierarchy is the key itself. Note, that the i18n package does not intentionally
// model any directories or nesting because it optimizes for an index lookup anyway.
type Path string

func (p Path) Parent() Path {
	if p == "" {
		return ""
	}

	dirs := i18n.Key(p).Directories()

	return Path(strings.Join(dirs, "."))
}

type Directory struct {
	Info        DirInfo
	Strings     []i18n.Key // immediate children which are not part of any nested children
	Directories []DirInfo  // children which are logically nested sections
}

func (d Directory) Identity() Path {
	return d.Info.Path
}

type DirInfo struct {
	Name             string // Name is a human-readable shorter representation of the path and is not unique.
	Path             Path
	TotalKeys        int // the total recursive amount of keys within the section
	TotalMissingKeys int // the total amount of recursive keys which are not yet translated in any way.
}

// FindResources returns the underlying reference resources instance. Usually this defaults to [i18n.Default].
type FindResources func(subject auth.Subject) (*i18n.Resources, error)

// ReadDir aggregates the flat space of all translatable strings into a recursive directory structure to support
// human centered inspection and translation processes.
type ReadDir func(subject auth.Subject, path Path) (Directory, error)

// ReadStringKeys returns an ordered slice of all those translatable string which are classified as StringKey
// and don't belong to any hierarchy. See also [i18n.Key.StringKey] for details.
type ReadStringKeys func(subject auth.Subject) ([]i18n.Key, error)

type UpdateMessage func(subject auth.Subject, tag language.Tag, msg i18n.Message) error

type AddLanguage func(subject auth.Subject, lang language.Tag) error

// Flush loads all updated strings, transfers them into [i18n.Default] and applies the flush.
type Flush func(subject auth.Subject) error

type UseCases struct {
	ReadDir        ReadDir
	ReadStringKeys ReadStringKeys
	FindResources  FindResources
	UpdateMessage  UpdateMessage
	Flush          Flush
	AddLanguage    AddLanguage
}

type StringData struct {
	Key       i18n.Key                      `json:"key"`
	Messages  map[language.Tag]i18n.Message `json:"messages"`
	UpdatedAt time.Time                     `json:"updatedAt"`
	UpdatedBy user.ID                       `json:"updatedBy"`
}

func (d StringData) Identity() i18n.Key {
	return d.Key
}

func (d StringData) String() string {
	return string(d.Key)
}

type Repository data.Repository[StringData, i18n.Key]

// NewUseCases performs a Flush during construction.
func NewUseCases(repo Repository, resources *i18n.Resources) (UseCases, error) {

	updated := 0
	for stringData, err := range repo.All() {
		if err != nil {
			return UseCases{}, err
		}

		for tag, message := range stringData.Messages {
			bnd, ok := resources.Bundle(tag)
			if !ok {
				bnd, _ = resources.AddLanguage(tag)
			}

			if err := bnd.Update(message); err != nil {
				return UseCases{}, fmt.Errorf("failed to update %v.%s: %w", tag, stringData.Key, err)
			}

			updated++
		}
	}

	resources.Flush()

	slog.Info("reconfigured localized messages", "updated", updated)

	return UseCases{
		ReadDir:        NewReadDir(resources),
		ReadStringKeys: NewReadStringKeys(resources),
		FindResources: func(subject auth.Subject) (*i18n.Resources, error) {
			return resources, nil
		},
		UpdateMessage: NewUpdateMessage(repo, resources),
		Flush: func(subject auth.Subject) error {
			resources.Flush()
			return nil
		},
		AddLanguage: NewAddLanguage(resources),
	}, nil
}
