// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"fmt"
	"iter"
	"log/slog"
	"reflect"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xjson"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

type TransitionType int

func (t TransitionType) String() string {
	switch t {
	case HumanInTheLoop:
		return "Human in the loop"
	default:
		return fmt.Sprintf("%d", int(t))
	}
}

const (
	HumanInTheLoop TransitionType = iota
)

type Transition struct {
	Type     TransitionType
	InEvent  reflect.Type
	OutEvent reflect.Type
}

type DeclareOptions struct {
	ID          ID // optional ID, leave empty to generate a new random id
	Name        string
	Description string
	Actions     []Action
	Transitions []Transition // additional non-functional transitions for documentation purposes only, e.g. for human-in-the-loop
}
type Declare func(subject user.Subject, opts DeclareOptions) (ID, error)

type FindDeclaredWorkflows func(subject user.Subject) iter.Seq2[DeclareOptions, error]

type FindDeclaredWorkflow func(subject user.Subject, id ID) (option.Opt[DeclareOptions], error)

// ProcessEvent takes the given type and tries to create, step or stop workflow instances. Note, that by
// definition workflow instances are prone to logical event races, e.g. if the system receives events multiple
// times, the actions are invoked concurrently (as by definition for event bus subscribers). This may also cause
// multiple stop events or multiple executions of actions. This is by design and if your workflow must exclude such
// behavior you have to ensure this in your workflow implementation e.g. by using (distributed) mutexes
// and/or an according persistent state. You cannot get non-blocking action execution and serialized
// behavior at the same time and the workflow engine wants to be as performant as possible, so that is our tradeoff.
type ProcessEvent func(subject user.Subject, evt any) error

type Analyzed struct {
	StartEvents      []Typename
	EventTypesByName map[Typename]reflect.Type
}

type Analyze func(subject user.Subject, id ID) (Analyzed, error)

type Event struct {
	Payload  any
	Instance Instance
	SavedAt  xtime.UnixMilliseconds
	SeqNo    int64
	ID       EventKey
}

func (e Event) Identity() EventKey {
	return e.ID
}

type RenderOptions struct {
	Language             language.Tag
	ShowEventFields      bool `label:"Zeige Ereignis-Felder"`
	ShowDescriptions     bool `label:"Zeige Beschreibungen als Kommentare"`
	ShowStereotypes      bool `label:"Zeige Stereotypen"`
	ShowExternalEventBus bool `label:"Zeige Ereignis-Bus"`
	ShowTitle            bool `label:"Zeige Titel"`
}
type Render func(subject user.Subject, wid ID, opts RenderOptions) (core.SVG, error)

// FindInstanceEvents can be used as a debug function to inspect the life of a workflow instance. It is also
// used to check if an instance was interrupted during execution and if an action call must continue.
type FindInstanceEvents func(subject user.Subject, id Instance) iter.Seq2[EventKey, error]

type FindInstanceEvent func(subject auth.Subject, key EventKey) (option.Opt[Event], error)

// SaveEvent just marshals the given event as-is and connects it with the given instance.
// It does not evaluate something.
type SaveEvent func(subject user.Subject, id Instance, evt any) error

type persistedEventData struct {
	SavedAt xtime.UnixMilliseconds `json:"t"`
	Payload xjson.AdjacentEnvelope `json:"p"`
}

type Status struct {
	State     InstanceState
	StartedAt time.Time
	StoppedAt time.Time
	Error     Error
	ID        Instance
}

type Error struct {
	Valid   bool
	Message string
	Type    Typename
}

func (s Status) Identity() Instance {
	return s.ID
}

type InstanceState int

func (s InstanceState) String() string {
	switch s {
	case InstanceRunning:
		return "aktiv"
	case InstanceDone:
		return "abgeschlossen"
	case InstanceNotFound:
		return "not found"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

const (
	InstanceNotFound InstanceState = iota
	InstanceRunning
	InstanceDone
)

type GetStatus func(subject auth.Subject, id Instance) (option.Opt[Status], error)

const RepositoryNameDeclaredWorkflows = "nago.workflow.declared"

type UseCases struct {
	Declare               Declare
	ProcessEvent          ProcessEvent
	Render                Render
	FindDeclaredWorkflows FindDeclaredWorkflows
	FindDeclaredWorkflow  FindDeclaredWorkflow
	CreateInstance        CreateInstance
	FindInstances         FindInstances
	Analyze               Analyze
	SaveEvent             SaveEvent
	FindInstanceEvents    FindInstanceEvents
	FindInstanceEvent     FindInstanceEvent
	GetStatus             GetStatus
}

func NewUseCases(bus events.Bus, instanceStore blob.Store, eventStore blob.Store) UseCases {
	var declarations concurrent.RWMap[ID, *workflow]

	createInstanceFn := NewCreateInstance(instanceStore)
	saveEventFn := NewSaveEvent(eventStore)
	getStatusFn := NewGetStatus(eventStore)
	findInstancesFn := NewFindInstances(instanceStore)
	processEventFn := NewProcessEvent(&declarations, createInstanceFn, saveEventFn, getStatusFn, findInstancesFn)

	bus.Subscribe(func(evt any) {
		if err := processEventFn(user.SU(), evt); err != nil {
			slog.Error("received event from bus but workflow engine cannot process it", "err", err.Error())
		}
	})

	return UseCases{
		Declare:               NewDeclare(&declarations),
		Render:                NewRender(&declarations),
		FindDeclaredWorkflows: NewFindDeclaredWorkflows(&declarations),
		FindDeclaredWorkflow:  NewFindDeclaredWorkflow(&declarations),
		FindInstances:         findInstancesFn,
		CreateInstance:        createInstanceFn,
		Analyze:               NewAnalyze(&declarations),
		ProcessEvent:          processEventFn,
		SaveEvent:             saveEventFn,
		FindInstanceEvents:    NewFindInstanceEvents(eventStore),
		FindInstanceEvent:     NewFindInstanceEvent(eventStore),
		GetStatus:             getStatusFn,
	}
}
