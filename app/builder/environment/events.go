// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package environment

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/app/builder/app"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xjson"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
)

type Event interface {
	event()
}

type NamespaceCreated struct {
	Name Ident `json:"name,omitempty"`
}

func (NamespaceCreated) event() {}

type NamespaceDeleted struct {
	Name Ident `json:"name,omitempty"`
}

func (NamespaceDeleted) event() {}

type AppIDUpdated struct {
	ID core.ApplicationID `json:"id,omitempty"`
}

func (AppIDUpdated) event() {}

type GitRepoUpdated struct {
	URL core.URI `json:"url,omitempty"`
}

func (GitRepoUpdated) event() {}

type StructCreated struct {
	Namespace Ident `json:"namespace,omitempty"`
	Name      Ident `json:"name,omitempty"`
}

func (StructCreated) event() {}

type TypeDeleted struct {
	Namespace Ident `json:"namespace,omitempty"`
	Name      Ident `json:"name,omitempty"`
}

func (TypeDeleted) event() {}

func init() {
	xjson.RegisterFor[NamespaceCreated]("NamespaceCreated")
	xjson.RegisterFor[AppIDUpdated]("AppIDUpdated")
	xjson.RegisterFor[GitRepoUpdated]("GitRepoUpdated")
	xjson.RegisterFor[NamespaceDeleted]("NamespaceDeleted")
	xjson.RegisterFor[StructCreated]("StructCreated")
	xjson.RegisterFor[TypeDeleted]("TypeDeleted")
}

//

type EventRepository data.Repository[EventBox, EID]

type EID string

var lastTS atomic.Int64
var tsMutex sync.Mutex

func NewID(prefix app.ID) EID {
	tsMutex.Lock()
	defer tsMutex.Unlock()

	var ts int64
	for {
		ts = time.Now().UnixMilli()
		if ts <= lastTS.Load() {
			time.Sleep(time.Millisecond)
		} else {
			break
		}
	}

	lastTS.Store(ts)
	return EID(fmt.Sprintf("%s_%014d", prefix, ts))
}

func (e EID) Split() (xtime.UnixMilliseconds, app.ID, bool) {
	tokens := strings.Split(string(e), "_")
	if len(tokens) != 2 {
		return 0, "", false
	}

	var ts xtime.UnixMilliseconds
	v, err := strconv.ParseInt(tokens[0], 10, 64)
	if err != nil {
		return 0, "", false
	}

	ts = xtime.UnixMilliseconds(v)
	return ts, app.ID(tokens[1]), true
}

type EventBox struct {
	ID        EID                    `json:"id"`
	CreatedBy user.ID                `json:"user"`
	Payload   xjson.AdjacentEnvelope `json:"payload"`
}

func NewEventBox(subject auth.Subject, app app.ID, evt Event) EventBox {
	if !xjson.Registered(reflect.TypeOf(evt)) {
		panic(fmt.Errorf("eventbox: invalid event type %T", evt))
	}

	box := EventBox{
		ID:        NewID(app),
		CreatedBy: subject.ID(),
		Payload:   xjson.NewAdjacentEnvelope(evt),
	}

	return box
}

func (e EventBox) Identity() EID {
	return e.ID
}

func (e EventBox) Unwrap() (Event, bool) {
	if e, ok := e.Payload.Value.(Event); ok {
		return e, true
	}

	return nil, false
}
