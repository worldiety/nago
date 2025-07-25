// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"context"
	"fmt"
	"github.com/worldiety/jsonptr"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/application/dataimport/parser"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std/concurrent"
	"io"
	"iter"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Key is a composite of <Batch-ID>/<unix milli>/<seq number>
type Key string

var seqNum atomic.Int64

func NewKey(batch SID) Key {
	return Key(fmt.Sprintf("%s/%13d/%05d", batch, time.Now().UnixMilli(), seqNum.Add(1)))
}

// Entry represents at least a parsed In-Object which belongs to a batch. The Transformed field is optional and may
// contain manually filled or edited fields which are passed to the importer. Manually means e.g. from a user or
// from an LLM.
type Entry struct {
	ID Key `json:"id,omitempty"`

	// In represents the parsed data, which may originate from e.g. a CSV or JSON file or PDF.
	// Never nil.
	In *jsonptr.Obj `json:"in,omitempty"`

	// Transformed can be nil and contains an already processed result which already
	// must match the typed target struct of the importer.
	Transformed *jsonptr.Obj `json:"out,omitempty"`

	// Confirmed is a flag which has been set by the user for his own orientation. This flag has no meaning
	// for the import process.
	Confirmed bool `json:"confirmed,omitempty"`

	// Ignore is a flag which tells the importer to ignore this entry.
	Ignored bool `json:"rejected,omitempty"`

	// Imported is a flag that tells the importer to ignore this entry, because it has already been imported.
	Imported      bool      `json:"imported,omitempty"`
	ImportedAt    time.Time `json:"importedAt,omitempty"`
	ImportedError string    `json:"importedError,omitempty"`

	LastModBy user.ID   `json:"lastModBy,omitempty"`
	LastModAt time.Time `json:"lastModAt,omitempty"`
}

// Transform either returns Transformed if not nil or applies the given transformation
// on In and returns the result.
func (e Entry) Transform(transformation Transformation) *jsonptr.Obj {
	if e.Transformed != nil {
		return e.Transformed
	}

	obj := &jsonptr.Obj{}
	for _, rule := range transformation.CopyRules {
		val, err := jsonptr.Eval(e.In, rule.SrcKey)
		if err != nil {
			// TODO specify error types, so that we can be more precise
			//slog.Error("failed to evaluate json pointer from importer transformation","err",err.Error(),"ptr",rule.SrcKey)
			continue
		}

		if rule.DstKey == "" {
			rule.DstKey = rule.SrcKey
		}

		if err := jsonptr.Put(obj, rule.DstKey, val); err != nil {
			slog.Error("failed to put rule to object", "key", rule.DstKey, "err", err.Error())
		}
	}

	return obj
}

func (e Entry) Identity() Key {
	return e.ID
}

type FindEntryByID func(subject auth.Subject, id Key) (option.Opt[Entry], error)

type UpdateEntryTransformation func(subject auth.Subject, id Key, t Transformation) error

type UpdateEntryConfirmation func(subject auth.Subject, id Key, confirmed bool) error

type UpdateEntryIgnored func(subject auth.Subject, id Key, ignored bool) error

type UpdateEntryTransformed func(subject auth.Subject, id Key, transformed *jsonptr.Obj) error

type EntryRepository data.Repository[Entry, Key]

// SID is an ID of a Batch.
type SID string

// A Staging represents an intermediate collection of parsed objects which are ready to import.
type Staging struct {
	ID        SID         `json:"id,omitempty"`
	CreatedAt time.Time   `json:"createdAt"`
	CreatedBy user.ID     `json:"createdBy,omitempty"`
	Name      string      `json:"name,omitempty"`
	Comment   string      `json:"comment,omitempty"`
	Importer  importer.ID `json:"importer,omitempty"`
	// Transformation is a deep copy, which may refer to an existing template or not.
	// However, we treat it always as an independent copy, to simplify one-shot imports
	// which don't need an extra stored and referenced Transformation.
	Transformation Transformation `json:"transformation,omitempty"`
}

func (b Staging) Identity() SID {
	return b.ID
}

type StagingCreationData struct {
	Name     string
	Comment  string
	Importer importer.ID
}
type CreateStaging func(subject auth.Subject, cdata StagingCreationData) (Staging, error)

type FindStagingByID func(subject auth.Subject, id SID) (option.Opt[Staging], error)
type DeleteStaging func(subject auth.Subject, staging SID) error
type FindStagingsForImporter func(subject auth.Subject, id importer.ID) iter.Seq2[Staging, error]

type StagingRepository = data.Repository[Staging, SID]

type CopyRule struct {
	SrcKey jsonptr.Ptr `json:"srcKey"`
	DstKey jsonptr.Ptr `json:"dstKey"`
}

func (r CopyRule) Apply(dst, src *jsonptr.Obj) error {
	srcVal, err := jsonptr.Eval(src, r.SrcKey)
	if err != nil {
		return err
	}

	if err := jsonptr.Put(dst, r.DstKey, srcVal); err != nil {
		return err
	}

	return nil
}

// TID is the ID of a Transformation.
type TID string

// A Transformation represents a rule set to define how to transform from a Src [jsonptr.Obj] into a Dst [jsonptr.Obj].
// Such a Transformation is intended to be re-used between different import [Batch]'es. This is usually
// specified by the (end-) user, who may import the same kind of data from different (formatted) sources, e.g.
// historically different CSV or PDF documents.
type Transformation struct {
	ID        TID        `json:"id,omitempty"`
	CopyRules []CopyRule `json:"copyRules"`
}

func (t Transformation) RuleBySrc(ptr jsonptr.Ptr) (CopyRule, bool) {
	for _, rule := range t.CopyRules {
		if rule.SrcKey == ptr {
			return rule, true
		}
	}

	return CopyRule{}, false
}

func (t Transformation) RuleByDst(ptr jsonptr.Ptr) (CopyRule, bool) {
	for _, rule := range t.CopyRules {
		if rule.DstKey == ptr {
			return rule, true
		}
	}

	return CopyRule{}, false
}

func (t Transformation) Identity() TID {
	return t.ID
}

type UpdateStagingTransformation func(subject auth.Subject, stage SID, transform Transformation) error

type FindImporters func(subject auth.Subject) iter.Seq2[importer.Importer, error]
type FindImporterByID func(subject auth.Subject, id importer.ID) (option.Opt[importer.Importer], error)
type FindParsers func(subject auth.Subject) iter.Seq2[parser.Parser, error]

type RegisterImporter func(subject auth.Subject, imp importer.Importer) error
type RegisterParser func(subject auth.Subject, p parser.Parser) error

type ParseStats struct {
	// Count is the amount of successfully parsed and stored entries in the staging.
	Count int64
}
type Parse func(subject auth.Subject, dst SID, src parser.ID, opts parser.Options, reader io.Reader) (ParseStats, error)

type ImportOptions struct {
	Context         context.Context
	ImporterOptions importer.Options
}
type Import func(subject auth.Subject, stage SID, dst importer.ID, opts ImportOptions) error

type Validate func(subject auth.Subject, key Key, imp importer.ID) error

// FindMatches returns the next top 3 matches from the data set. The first Match has the highest score.
type FindMatches func(subject auth.Subject, key Key, imp importer.ID) ([]importer.Match, error)

type StagingReviewStatus struct {
	Total     int
	Confirmed int // confirmed also contains imported entries
	Ignored   int
	Imported  int

	// The following fields are only set, if [CalculateStagingReviewStatusOptions.Position] has been declared.

	PreviousEntry Key
	CurrentEntry  Key
	NextEntry     Key
}

func (s StagingReviewStatus) Checked() int {
	return s.Confirmed + s.Ignored
}

type CalculateStagingReviewStatusOptions struct {
	Position Key
}

type CalculateStagingReviewStatus func(subject auth.Subject, staging SID, opts CalculateStagingReviewStatusOptions) (StagingReviewStatus, error)

type FilterEntries func(subject auth.Subject, stage SID, opts data.PaginateOptions) (data.Page[Entry], error)

type UseCases struct {
	RegisterImporter             RegisterImporter
	RegisterParser               RegisterParser
	FindImporters                FindImporters
	FindParsers                  FindParsers
	FindStagingsForImporter      FindStagingsForImporter
	FindStagingByID              FindStagingByID
	FindImporterByID             FindImporterByID
	Parse                        Parse
	CreateStaging                CreateStaging
	FilterEntries                FilterEntries
	DeleteStaging                DeleteStaging
	UpdateStagingTransformation  UpdateStagingTransformation
	FindEntryByID                FindEntryByID
	UpdateEntryConfirmation      UpdateEntryConfirmation
	UpdateEntryIgnored           UpdateEntryIgnored
	UpdateEntryTransformation    UpdateEntryTransformation
	CalculateStagingReviewStatus CalculateStagingReviewStatus
	UpdateEntryTransformed       UpdateEntryTransformed
	Import                       Import
}

func NewUseCases(repoStaging StagingRepository, repoEntry EntryRepository) UseCases {
	var parsers concurrent.RWMap[parser.ID, parser.Parser]
	var imports concurrent.RWMap[importer.ID, importer.Importer]

	var mutex sync.Mutex
	return UseCases{
		RegisterImporter:             NewRegisterImporter(&imports),
		RegisterParser:               NewRegisterParser(&parsers),
		FindImporters:                NewFindImporters(&imports),
		FindParsers:                  NewFindParsers(&parsers),
		FindStagingsForImporter:      NewFindStagingsForImporter(repoStaging),
		FindStagingByID:              NewFindStagingByID(repoStaging),
		FindImporterByID:             NewFindImporterByID(&imports),
		Parse:                        NewParse(repoStaging, repoEntry, &parsers),
		CreateStaging:                NewCreateStaging(&mutex, repoStaging),
		FilterEntries:                NewFilterEntries(repoEntry),
		DeleteStaging:                NewDeleteStaging(repoStaging, repoEntry),
		UpdateStagingTransformation:  NewUpdateStagingTransformation(&mutex, repoStaging),
		FindEntryByID:                NewFindEntryByID(repoEntry),
		UpdateEntryIgnored:           NewUpdateEntryIgnored(&mutex, repoEntry),
		UpdateEntryConfirmation:      NewUpdateEntryConfirmation(&mutex, repoEntry),
		CalculateStagingReviewStatus: NewCalculateStagingReviewStatus(repoStaging, repoEntry),
		UpdateEntryTransformed:       NewUpdateEntryTransformed(&mutex, repoEntry),
		Import:                       NewImport(&mutex, repoEntry, repoStaging, &imports),
	}
}
