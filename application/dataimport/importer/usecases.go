// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package importer

import (
	"context"
	"iter"
	"reflect"

	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/presentation/core"
)

type ID string

type Configuration struct {
	Image       core.SVG
	Name        string
	Description string

	// Passthrough indicates that this importer does not need any transformation rules and instead
	// should get passed the parsed raw data.
	Passthrough bool

	// ExpectedType represents the actual type which is used to coerce the schema of the [jsonptr.Obj].
	ExpectedType reflect.Type

	ImportOptionsType reflect.Type

	// PreviewMappings are evaluated to generate a preview from input or transformed data.
	// If there are dozens of fields, we cannot display all of them. And usually a lot of them are
	// not significant. Thus give some display hints.
	PreviewMappings []PreviewMapping
}

type PreviewMapping struct {
	Name string
	// Keywords either are matching as a substring or if started with / as a json pointer. A json pointer is
	// evaluated case sensitive. Other strings must be lower case and are evaluated that way.
	Keywords []string
}

type Options struct {
	// ContinueOnError tells the Importer to import as much as entries it can.
	ContinueOnError bool

	// MergeDuplicates tells the importer, that candidates which are almost certain a duplicate, shall be merged
	// if a unique constraint of an identity would otherwise fail.
	MergeDuplicates bool

	// Options is either nil or contains an instance of [Configuration.ImportOptionsType].
	Options any
}

type MatchOptions struct {
}

type Importer interface {
	Identity() ID
	Configuration() Configuration
	Import(ctx context.Context, opts Options, data iter.Seq2[*jsonptr.Obj, error]) error

	// Validate takes the given object and validates it as if it would be imported.
	Validate(ctx context.Context, obj *jsonptr.Obj) error

	// FindMatches calculates similarity distances between the given object and all other already stored
	// entities. An implementation may but must not return matches with a score of 0. The resulting sequence is
	// unordered.
	FindMatches(ctx context.Context, opts MatchOptions, obj *jsonptr.Obj) iter.Seq2[Match, error]
}

type Match struct {
	// Obj contains the converted duplicate candidate.
	Obj *jsonptr.Obj

	// A Score of 1 is 100% duplicate. A Score of 0 is definitely not a duplicate.
	Score float64
}
