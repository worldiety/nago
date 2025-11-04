// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package libsync

import (
	"fmt"
	"iter"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/drive"
	"go.wdy.de/nago/application/ent"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xsync"
)

type Source struct {
	Store struct {
		Valid bool   `json:"valid,omitempty"`
		Name  string `json:"name,omitempty"`
	} `json:"store"`

	Drive struct {
		Valid bool      `json:"valid,omitempty"`
		Root  drive.FID `json:"root,omitempty"`
	} `json:"drive"`
}

type Job struct {
	ID            library.ID    `json:"id,omitempty"`
	Provider      provider.ID   `json:"provider,omitempty"`
	Sources       []Source      `json:"sources,omitempty"`
	PullPauseTime time.Duration `json:"pullPauseTime,omitempty"`
}

func (e Job) WithIdentity(id library.ID) Job {
	e.ID = id
	return e
}

func (e Job) Identity() library.ID {
	return e.ID
}

type SourceDocument struct {
	Store struct {
		Valid bool   `json:"valid"`
		Name  string `json:"name"`
		Key   string `json:"key"`
	} `json:"store"`

	Drive struct {
		Valid bool      `json:"valid"`
		File  drive.FID `json:"file"`
	} `json:"drive"`
}

type SyncInfo struct {
	// Remote represents the library document. Note, that the backends do not support updating a document.
	// Instead, it must be removed and inserted again.
	Remote document.ID `json:"remote"`

	Src  SourceDocument `json:"src"`
	Size int64          `json:"size"`

	// Hash is calculated locally and represents the value hash of the Store/Key entry which has been uploaded
	// and is represented by the Remote document id.
	Hash string `json:"hash"`
}

func (e SyncInfo) Identity() document.ID {
	return e.Remote
}

// Create inserts a new sync job for the library. There is at most a single Job per library.
type Create func(subject auth.Subject, job Job) (library.ID, error)

// DeleteByID removes the entire job.
type DeleteByID func(subject auth.Subject, job library.ID) error

type Synchronize func(subject auth.Subject, lib library.ID) error

type FindAll func(subject auth.Subject) iter.Seq2[Job, error]

type FindByID func(subject auth.Subject, job library.ID) (option.Opt[Job], error)

type FindAllIdentifiers func(subject auth.Subject) iter.Seq2[library.ID, error]

type AddSource func(subject auth.Subject, job library.ID, src Source) error
type RemoveSource func(subject auth.Subject, job library.ID, src Source) error

type Update func(subject auth.Subject, job Job) error

type Repository data.Repository[Job, library.ID]

type SyncRepository data.Repository[SyncInfo, document.ID]

type UseCases struct {
	Delete             DeleteByID
	Create             Create
	FindAll            FindAll
	FindByID           FindByID
	FindAllIdentifiers FindAllIdentifiers
	Update             Update
	Synchronize        Synchronize
	AddSource          AddSource
	RemoveSource       RemoveSource
}

// NewUseCases also listens for the following events on the event bus, to trigger synchronizations automatically:
//   - [drive.Activity]
//   - [blob.Written]
//   - [blob.Deleted]
func NewUseCases(bus events.Bus, findProvider ai.FindProviderByID, jobRepo Repository, syncRepo SyncRepository, stores blob.Stores, walkDir drive.WalkDir, openFile drive.Get, statFile drive.Stat) UseCases {
	var mutex sync.Mutex
	uc := ent.NewUseCases(Permissions, jobRepo, ent.Options{Mutex: &mutex})

	syncFn := NewSynchronize(bus, findProvider, jobRepo, syncRepo, stores, walkDir, openFile, statFile)

	var lastMod atomic.Int64
	xsync.GoFn(
		func() {
			// TODO do we need something to stop? a context?
			// TODO we need an infrastructure in NAGO to notify about errors which otherwise would go unnoticed
			var lastModProcess int64
			for range time.Tick(time.Minute * 1) {
				if lastModProcess == lastMod.Load() {
					continue
				}

				// TODO what about unsupported file types? the api will reject that???
				for info, err := range jobRepo.All() {
					if err != nil {
						slog.Error("failed to execute job repo listing", "err", err.Error())
						continue
					}

					if err := syncFn(user.SU(), info.ID); err != nil {
						slog.Error("failed to execute libsync", "err", err.Error())
						continue
					}
				}

				lastModProcess = lastMod.Load() // TODO not clear if we stop syncing on error or should retry in next cycle: e.g. if unsupported mimetype, we would run in a cycle?
			}
		},
	)

	bus.Subscribe(func(evt any) {
		switch evt.(type) {
		case drive.Activity:
			lastMod.Add(1)
		case blob.Written:
			lastMod.Add(1)
		case blob.Deleted:
			lastMod.Add(1)
		}
	})

	return UseCases{
		Synchronize: syncFn,
		Delete:      DeleteByID(uc.DeleteByID),
		Create: func(subject auth.Subject, job Job) (library.ID, error) {
			if job.ID == "" {
				return "", fmt.Errorf("job id must exist and match library id")
			}

			return uc.Create(subject, job)
		},
		FindAll:            FindAll(uc.FindAll),
		FindAllIdentifiers: FindAllIdentifiers(uc.FindAllIdentifiers),
		Update:             Update(uc.Update),
		FindByID:           FindByID(uc.FindByID),
		AddSource:          NewAddSource(&mutex, jobRepo),
		RemoveSource:       NewRemoveSource(&mutex, jobRepo),
	}
}
