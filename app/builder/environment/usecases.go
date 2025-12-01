// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Ident: Custom-License

package environment

import (
	"iter"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/app/builder/app"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

type ID string

type Environment struct {
	ID          ID        `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Owner       []user.ID `json:"owner"`
	Apps        []app.ID  `json:"apps"`
}

func (e Environment) Identity() ID {
	return e.ID
}

type CreateOptions struct {
	Name        string `label:"nago.common.label.name"`
	Description string `label:"nago.common.label.description" lines:"3"`
}

type Create func(subject auth.Subject, opts CreateOptions) (ID, error)

type FindAll func(subject auth.Subject) iter.Seq2[Environment, error]

type FindAppByID func(subject auth.Subject, env ID, id app.ID) (option.Opt[app.App], error)

type CreateAppOptions struct {
	Name        string `label:"nago.common.label.name"`
	Description string `label:"nago.common.label.description" lines:"3"`
}
type CreateApp func(subject auth.Subject, env ID, opts CreateAppOptions) (app.ID, error)

type PutEvent func(subject auth.Subject, env ID, app app.ID, event Event) error

type Replay func(subject auth.Subject, env ID, app app.ID) iter.Seq2[EventBox, error]

type UseCases struct {
	Create      Create
	FindAll     FindAll
	FindAppByID FindAppByID
	CreateApp   CreateApp
	PutEvent    PutEvent
	Replay      Replay
}

type Repository data.Repository[Environment, ID]

func NewUseCases(envRepo Repository, appRepo app.Repository, evtRepo EventRepository) UseCases {
	var mutex sync.Mutex

	findAppByIDFn := NewFindAppByID(envRepo, appRepo)
	return UseCases{
		Create:      NewCreate(&mutex, envRepo),
		FindAll:     NewFindAll(envRepo),
		FindAppByID: findAppByIDFn,
		CreateApp:   NewCreateApp(&mutex, envRepo, appRepo),
		PutEvent:    NewPutEvent(evtRepo, findAppByIDFn),
		Replay:      NewReplay(evtRepo, findAppByIDFn),
	}
}
