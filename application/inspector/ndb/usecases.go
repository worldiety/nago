// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndbinspector

import (
	"fmt"
	"slices"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

// msgstoreDB is the accessor a msgstore-backed engine exposes to reach its
// concrete database (see msgstore/engine.go). We type-assert to it rather than
// importing the unexported engine type.
type msgstoreDB interface {
	DB() *msgstore.DB
}

// Instance is one ndb database registered with the application, identified by a
// stable Path and labeled by Name.
type Instance struct {
	Path string
	Name string
	DB   *ndb.DB
}

// InstancesProvider returns all currently registered ndb databases. It is called
// on each use-case invocation so databases opened after startup are included.
type InstancesProvider func() []Instance

// InstanceRef identifies one ndb database in the UI (by its path).
type InstanceRef struct {
	Path string
	Name string
}

func (r InstanceRef) Identity() string { return r.Path }

// EngineRef identifies one engine instance inside an ndb database.
type EngineRef struct {
	Instance string // instance path
	Name     string
	Kind     ndb.EngineKind
}

func (e EngineRef) Identity() string { return e.Instance + "|" + e.Name }

// TypeInfo is the metadata-only description of one message stream (event type).
type TypeInfo struct {
	Type       ndb.TypeID
	Segments   int
	Bytes      int64
	MinSeq     ndb.Seq
	MaxSeq     ndb.Seq
	HasPending bool
}

func (t TypeInfo) Identity() string { return string(t.Type) }

// MessageRow is one message rendered into a table row. Payload is a copy safe to
// retain (Replay yields reusable buffers; the use case clones them).
type MessageRow struct {
	Type     ndb.TypeID
	Seq      ndb.Seq
	TimeNano int64
	TraceID  string
	Encoding uint8
	Tomb     bool
	Size     int
	Payload  []byte
}

func (m MessageRow) Identity() string { return fmt.Sprintf("%d", m.Seq) }

// WindowRequest bounds a single replay window so the UI never materializes a
// whole stream. At most Limit rows are returned starting at MinSeq.
//
// Types may contain one or more message types. When several are given, the
// engine performs a k-way merge and the returned rows are coherent in global
// Seq order across all selected types (so the list stays sequence-ordered even
// with a mixed selection). An empty Types slice selects all types.
type WindowRequest struct {
	Instance string // instance path
	Engine   string
	Types    []ndb.TypeID
	MinSeq   ndb.Seq
	MaxSeq   ndb.Seq
	Limit    int // 0 -> DefaultWindowLimit
}

// DefaultWindowLimit is the hard cap on rows returned per replay window.
const DefaultWindowLimit = 200

// UseCases is the ndb inspector domain surface across all registered ndb
// databases. All operations are guarded by PermNDBInspector.
type UseCases struct {
	instances InstancesProvider
}

// NewUseCases builds the inspector use cases over a live provider of ndb
// databases, so every database registered with the application is inspectable.
func NewUseCases(instances InstancesProvider) UseCases {
	return UseCases{instances: instances}
}

// Instances lists the registered ndb databases, sorted by path.
func (uc UseCases) Instances(subject auth.Subject) ([]InstanceRef, error) {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return nil, err
	}
	all := uc.instances()
	out := make([]InstanceRef, 0, len(all))
	for _, in := range all {
		out = append(out, InstanceRef{Path: in.Path, Name: in.Name})
	}
	slices.SortFunc(out, func(a, b InstanceRef) int { return cmpStr(a.Path, b.Path) })
	return out, nil
}

// dbByPath resolves the ndb database for an instance path.
func (uc UseCases) dbByPath(path string) (*ndb.DB, error) {
	for _, in := range uc.instances() {
		if in.Path == path {
			return in.DB, nil
		}
	}
	return nil, fmt.Errorf("ndb instance %q not found", path)
}

// MessageEngines lists the msgstore engine instances of one ndb database.
func (uc UseCases) MessageEngines(subject auth.Subject, instancePath string) ([]EngineRef, error) {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return nil, err
	}
	db, err := uc.dbByPath(instancePath)
	if err != nil {
		return nil, err
	}
	var out []EngineRef
	for info, err := range db.Engines() {
		if err != nil {
			return nil, err
		}
		if info.Kind != msgstore.EngineKind {
			continue
		}
		out = append(out, EngineRef{Instance: instancePath, Name: info.Name, Kind: info.Kind})
	}
	slices.SortFunc(out, func(a, b EngineRef) int { return cmpStr(a.Name, b.Name) })
	return out, nil
}

// messages resolves the concrete msgstore DB for the named engine in an instance.
func (uc UseCases) messages(instancePath, engine string) (*msgstore.DB, error) {
	db, err := uc.dbByPath(instancePath)
	if err != nil {
		return nil, err
	}
	opt, err := db.LookupEngine(engine)
	if err != nil {
		return nil, err
	}
	if opt.IsNone() {
		return nil, fmt.Errorf("engine %q not found", engine)
	}
	acc, ok := opt.Unwrap().(msgstoreDB)
	if !ok {
		return nil, fmt.Errorf("engine %q is not a msgstore engine", engine)
	}
	return acc.DB(), nil
}

// Types lists the message streams of an engine with cheap, metadata-only stats.
func (uc UseCases) Types(subject auth.Subject, instancePath, engine string) ([]TypeInfo, error) {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return nil, err
	}
	db, err := uc.messages(instancePath, engine)
	if err != nil {
		return nil, err
	}
	types, err := db.Types()
	if err != nil {
		return nil, err
	}
	out := make([]TypeInfo, 0, len(types))
	for _, t := range types {
		st, err := db.TypeStat(t)
		if err != nil {
			return nil, err
		}
		out = append(out, TypeInfo{
			Type: st.Type, Segments: st.Segments, Bytes: st.Bytes,
			MinSeq: st.MinSeq, MaxSeq: st.MaxSeq, HasPending: st.HasPending,
		})
	}
	return out, nil
}

// SeqForTime resolves the smallest Seq whose append time is >= tsNano, to seek a
// replay window by wall-clock time.
func (uc UseCases) SeqForTime(subject auth.Subject, instancePath, engine string, tsNano int64) (ndb.Seq, error) {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return 0, err
	}
	db, err := uc.messages(instancePath, engine)
	if err != nil {
		return 0, err
	}
	return db.SeqForTime(tsNano)
}

// Window replays a single bounded window of messages for the selected types. It
// returns at most min(req.Limit, DefaultWindowLimit) rows in ascending global
// Seq order (a k-way merge across all selected types, so the list is coherent
// even for a mixed selection), each with a cloned payload. This is the only read
// path — the UI pages by moving MinSeq.
func (uc UseCases) Window(subject auth.Subject, req WindowRequest) ([]MessageRow, error) {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return nil, err
	}
	db, err := uc.messages(req.Instance, req.Engine)
	if err != nil {
		return nil, err
	}
	limit := req.Limit
	if limit <= 0 || limit > DefaultWindowLimit {
		limit = DefaultWindowLimit
	}
	maxSeq := req.MaxSeq
	if maxSeq == 0 {
		maxSeq = ndb.Seq(^uint64(0))
	}

	var rows []MessageRow
	for _, msg := range db.Replay(req.Types, req.MinSeq, maxSeq) {
		rows = append(rows, MessageRow{
			Type:     msg.Type,
			Seq:      msg.Seq,
			TimeNano: msg.TimeNano,
			TraceID:  msg.TraceID.String(),
			Encoding: uint8(msg.Encoding),
			Tomb:     msg.IsTombstone(),
			Size:     len(msg.Payload),
			Payload:  slices.Clone(msg.Payload),
		})
		if len(rows) >= limit {
			break
		}
	}
	return rows, nil
}

// DeleteSeq soft-deletes (tombstones) a single message. Knife tool.
func (uc UseCases) DeleteSeq(subject auth.Subject, instancePath, engine string, typeID ndb.TypeID, seq ndb.Seq) error {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return err
	}
	db, err := uc.messages(instancePath, engine)
	if err != nil {
		return err
	}
	return db.DeleteSeq(typeID, seq)
}

// DeleteType removes an entire message stream. Knife tool.
func (uc UseCases) DeleteType(subject auth.Subject, instancePath, engine string, typeID ndb.TypeID) error {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return err
	}
	db, err := uc.messages(instancePath, engine)
	if err != nil {
		return err
	}
	return db.DeleteType(typeID)
}

// RebuildTimeIndex rebuilds the time index of an engine. Knife tool; the caller
// must ensure no concurrent writes.
func (uc UseCases) RebuildTimeIndex(subject auth.Subject, instancePath, engine string) error {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return err
	}
	db, err := uc.messages(instancePath, engine)
	if err != nil {
		return err
	}
	return db.RebuildTimeIndex()
}

func cmpStr(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
