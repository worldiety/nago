// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"iter"
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"sync"

	enumjson "github.com/worldiety/enum/json"
	"go.wdy.de/nago/application/backup"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
)

// Heads keeps the heads of the sessions (see [Summary]) in store, so that [UseCases.FindResumable], the history
// ([UseCases.FindSummaries]), [UseCases.FindAll] with [FindAllOptions.CreatedBy] and [UseCases.Delete] read the heads
// of a person or the children of a session instead of every session. The store must be opened with
// [blob.OpenStoreOptions.Derived], ideally in the same database as the sessions, so that both share one log.
//
// The use cases write every session together with its head. They build the heads of the sessions in the background,
// once for sessions which never had heads, after a [backup.Restored] on the event bus of ctx, if ctx carries one like
// the context of the configurator, and at the start after an interrupted build or restore. Until a build has finished,
// the searches read the sessions it has not reached yet, and while the heads are invalid, every session. ctx bounds
// the builds. Without Heads, every use case reads every session.
//
// The use cases of the process share the locks and the build of a store, if they are constructed with the same store.
// Sessions written by anything else but these use cases, e.g. the data inspector, appear only after the next build;
// a head, whose session is gone, does not count.
func Heads(ctx context.Context, store blob.Store) Option {
	return func(o *useCaseOptions) {
		o.headsCtx, o.heads = ctx, store
	}
}

// The keys of a store of heads.
const (
	// keyHeads + creator + "/w/" + id is the head of a session without parent, which waits on a decision, and
	// keyHeads + creator + "/n/" + id the head of any other session without parent.
	keyHeads = "h/"
	// keyChildren + parent + "/" + id is the head of a child session.
	keyChildren = "k/"
	// keyMarks + id marks a session with an open write, see heads.save.
	keyMarks = "o/"
	// keyLayout holds the layout.
	keyLayout = "~layout"
)

// layout is the state of the build of the heads.
type layout struct {
	Version int    `json:"version"`
	State   string `json:"state"`
	// Cursor is the last session, which the running build has derived. The build continues behind it.
	Cursor ID `json:"cursor,omitempty"`
	// Skipped counts the sessions, which the build could not decode.
	Skipped int `json:"skipped,omitempty"`
}

const (
	layoutVersion = 1
	stateBuilding = "building"
	stateDone     = "done"
)

// mark is the value of keyMarks: what locates the heads of the session, whose content may be gone.
type mark struct {
	CreatedBy user.ID `json:"createdBy,omitempty"`
	ParentID  ID      `json:"parentId,omitempty"`
}

// heads keeps the heads of the sessions of repo in store, see [Heads].
//
// Every write of a session runs under one of the locks of the store in this order: mark the session as open, write
// it, write its head and remove its head under the other state, remove the mark. A process ends at any point of it
// with a prefix of these writes, because the store applies them in their order. Thus, where a mark exists, the
// session itself decides and its head is ignored, otherwise the head. A head, whose session does not exist, does not
// count either.
type heads struct {
	ctx   context.Context
	repo  Repository // the sessions themselves
	store blob.Store
}

func newHeads(ctx context.Context, repo Repository, store blob.Store) *heads {
	if ctx == nil {
		ctx = context.Background()
	}

	return &heads{ctx: ctx, repo: repo, store: store}
}

// headLocks is the number of locks per store of heads.
const headLocks = 256

// headShare is what the use cases of the process share for a store of heads: the locks of the writes and the build.
// It exists, while one of its locks is held or a build runs, so that no store stays referenced afterwards.
type headShare struct {
	refs  int
	locks [headLocks]sync.Mutex
	// building tells, whether a build runs, and again, whether another one has been asked for meanwhile
	building, again bool
}

var headShares = struct {
	mu sync.Mutex
	m  map[any]*headShare
}{m: map[any]*headShare{}}

// shareKey identifies the store within the process: by itself, if it can be compared, otherwise by its name.
func shareKey(store blob.Store) any {
	if reflect.ValueOf(store).Comparable() {
		return store
	}

	return store.Name()
}

// acquireShare returns the share of the key with one more reference. The caller holds headShares.mu.
func acquireShare(key any) *headShare {
	sh := headShares.m[key]
	if sh == nil {
		sh = &headShare{}
		headShares.m[key] = sh
	}

	sh.refs++
	return sh
}

// releaseShare drops a reference of the share. The caller holds headShares.mu.
func releaseShare(key any, sh *headShare) {
	sh.refs--
	if sh.refs == 0 {
		delete(headShares.m, key)
	}
}

// lock acquires the lock of the session id in the store and returns its release.
func (h *heads) lock(id ID) (release func()) {
	key := shareKey(h.store)
	headShares.mu.Lock()
	sh := acquireShare(key)
	headShares.mu.Unlock()

	hash := fnv.New32a()
	_, _ = hash.Write([]byte(id))
	mu := &sh.locks[hash.Sum32()%headLocks]
	mu.Lock()

	return func() {
		mu.Unlock()
		headShares.mu.Lock()
		releaseShare(key, sh)
		headShares.mu.Unlock()
	}
}

// headKey is the key of the head of the session.
func headKey(s Session) string {
	if s.ParentID != "" {
		return keyChildren + string(s.ParentID) + "/" + string(s.ID)
	}

	if s.Pending != nil {
		return keyHeads + string(s.CreatedBy) + "/w/" + string(s.ID)
	}

	return keyHeads + string(s.CreatedBy) + "/n/" + string(s.ID)
}

// headKeys returns every key, under which a head of the session may be.
func headKeys(id ID, createdBy user.ID, parent ID) []string {
	keys := []string{
		keyHeads + string(createdBy) + "/w/" + string(id),
		keyHeads + string(createdBy) + "/n/" + string(id),
	}

	if parent != "" {
		keys = append(keys, keyChildren+string(parent)+"/"+string(id))
	}

	return keys
}

// idOfKey returns the session of a key of a head.
func idOfKey(key string) ID {
	return ID(key[strings.LastIndexByte(key, '/')+1:])
}

func (h *heads) get(key string, v any) (bool, error) {
	opt, err := blob.Get(h.store, key)
	if err != nil {
		return false, err
	}

	if opt.IsNone() {
		return false, nil
	}

	if err := stdjson.Unmarshal(opt.Unwrap(), v); err != nil {
		return false, fmt.Errorf("cannot decode %s of %s: %w", key, h.store.Name(), err)
	}

	return true, nil
}

func (h *heads) put(key string, v any) error {
	buf, err := stdjson.Marshal(v)
	if err != nil {
		return err
	}

	return blob.Put(h.store, key, buf)
}

func (h *heads) del(key string) error {
	return h.store.Delete(context.Background(), key)
}

func (h *heads) exists(key string) (bool, error) {
	return h.store.Exists(context.Background(), key)
}

// keys returns the keys with the prefix.
func (h *heads) keys(prefix string) ([]string, error) {
	var res []string
	for key, err := range h.store.List(context.Background(), blob.ListOptions{Prefix: prefix}) {
		if err != nil {
			return nil, fmt.Errorf("cannot list %s of %s: %w", prefix, h.store.Name(), err)
		}

		res = append(res, key)
	}

	return res, nil
}

// contentExists reports whether the session exists, without reading it.
func (h *heads) contentExists(id ID) (bool, error) {
	for got, err := range h.repo.IdentifiersByPrefix(id) {
		if err != nil {
			return false, err
		}

		// the id itself is the smallest key with this prefix
		return got == id, nil
	}

	return false, nil
}

// save writes the session and its head, see heads. A failure after the session has been written is only logged: the
// mark keeps the searches right, and the next start repairs the head.
func (h *heads) save(s Session) error {
	release := h.lock(s.ID)
	defer release()

	if err := h.put(keyMarks+string(s.ID), mark{CreatedBy: s.CreatedBy, ParentID: s.ParentID}); err != nil {
		return fmt.Errorf("cannot mark the session as written: %w", err)
	}

	if err := h.repo.Save(s); err != nil {
		return err
	}

	if err := h.settle(s); err != nil {
		slog.Error("cannot write the head of a session, its mark stays", "session", s.ID, "err", err)
		return nil
	}

	if err := h.del(keyMarks + string(s.ID)); err != nil {
		slog.Error("cannot remove the mark of a written session", "session", s.ID, "err", err)
	}

	return nil
}

// settle writes the head of the session and removes it under the other state. The caller holds the lock.
func (h *heads) settle(s Session) error {
	key := headKey(s)
	if err := h.put(key, s.Summary()); err != nil {
		return err
	}

	for _, other := range headKeys(s.ID, s.CreatedBy, s.ParentID) {
		if other == key {
			continue
		}

		if err := h.del(other); err != nil {
			return err
		}
	}

	return nil
}

// remove deletes the session and its heads, see heads.
func (h *heads) remove(s Session) error {
	release := h.lock(s.ID)
	defer release()

	if err := h.put(keyMarks+string(s.ID), mark{CreatedBy: s.CreatedBy, ParentID: s.ParentID}); err != nil {
		return fmt.Errorf("cannot mark the session as deleted: %w", err)
	}

	if err := h.repo.DeleteByID(s.ID); err != nil {
		return err
	}

	for _, key := range headKeys(s.ID, s.CreatedBy, s.ParentID) {
		if err := h.del(key); err != nil {
			slog.Error("cannot remove the head of a deleted session, its mark stays", "session", s.ID, "err", err)
			return nil
		}
	}

	if err := h.del(keyMarks + string(s.ID)); err != nil {
		slog.Error("cannot remove the mark of a deleted session", "session", s.ID, "err", err)
	}

	return nil
}

// coverage tells, which sessions the heads cover: all, those up to a cursor, or none.
type coverage struct {
	all    bool
	cursor ID
}

func (c coverage) covers(id ID) bool {
	return c.all || (c.cursor != "" && id <= c.cursor)
}

func (c coverage) none() bool {
	return !c.all && c.cursor == ""
}

// coverage returns what the heads cover now: nothing while they are invalid or before a build began, the sessions up
// to the cursor of a build and all after a build.
func (h *heads) coverage() (coverage, error) {
	invalid, err := h.exists(blob.InvalidKey)
	if err != nil {
		return coverage{}, err
	}

	if invalid {
		return coverage{}, nil
	}

	var l layout
	ok, err := h.get(keyLayout, &l)
	if err != nil || !ok {
		return coverage{}, err
	}

	if l.State == stateDone {
		return coverage{all: true}, nil
	}

	return coverage{cursor: l.Cursor}, nil
}

// marked returns the sessions with a mark, for which the session itself decides. A search reads it first, so that a
// write, which ends meanwhile, has written the head before.
func (h *heads) marked() (map[ID]bool, error) {
	keys, err := h.keys(keyMarks)
	if err != nil {
		return nil, err
	}

	res := make(map[ID]bool, len(keys))
	for _, key := range keys {
		res[ID(strings.TrimPrefix(key, keyMarks))] = true
	}

	return res, nil
}

// load reads a session for a search, which the heads have named or do not cover.
func (h *heads) load(id ID) (Session, bool, error) {
	opt, err := h.repo.FindByID(id)
	if err != nil {
		return Session{}, false, fmt.Errorf("cannot load session: %w", err)
	}

	if opt.IsNone() {
		return Session{}, false, nil
	}

	return opt.Unwrap(), true, nil
}

// uncovered yields the sessions which c does not cover, in the order of the ids.
func (h *heads) uncovered(c coverage) iter.Seq2[Session, error] {
	return func(yield func(Session, error) bool) {
		if c.all {
			return
		}

		for id, err := range h.repo.Identifiers() {
			if err != nil {
				if !yield(Session{}, err) {
					return
				}
				continue
			}

			if c.covers(id) {
				continue
			}

			s, ok, err := h.load(id)
			if err != nil {
				if !yield(Session{}, err) {
					return
				}
				continue
			}

			if ok && !yield(s, nil) {
				return
			}
		}
	}
}

// candidates returns the sessions of the heads with the prefixes, which c covers and which have no mark, sorted by id
// and without duplicates.
func (h *heads) candidates(c coverage, marked map[ID]bool, prefixes ...string) ([]ID, error) {
	var res []ID
	for _, prefix := range prefixes {
		keys, err := h.keys(prefix)
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			if id := idOfKey(key); c.covers(id) && !marked[id] {
				res = append(res, id)
			}
		}
	}

	slices.Sort(res)
	return slices.Compact(res), nil
}

// undecodable reports whether err tells, that a stored session has been read but cannot be decoded, unlike an error of
// reading it: a [json.DecodeError] of nago's JSON repository, or an error of a JSON decoder of another repository.
func undecodable(err error) bool {
	var (
		decode    *json.DecodeError
		syntax    *enumjson.SyntaxError
		typ       *enumjson.UnmarshalTypeError
		stdSyntax *stdjson.SyntaxError
		stdTyp    *stdjson.UnmarshalTypeError
	)

	return errors.As(err, &decode) || errors.As(err, &syntax) || errors.As(err, &typ) || errors.As(err, &stdSyntax) ||
		errors.As(err, &stdTyp)
}

// start repairs the heads of the sessions with a mark and starts the build, which the store needs, see [Heads].
func (h *heads) start() {
	h.repairMarks()

	invalid, err := h.exists(blob.InvalidKey)
	if err != nil {
		slog.Error("cannot tell whether the heads of the sessions are valid", "store", h.store.Name(), "err", err)
		return
	}

	if invalid {
		// a restore did not end, or its end has not been seen
		h.kick()
	} else {
		var l layout
		ok, err := h.get(keyLayout, &l)
		switch {
		case err != nil:
			slog.Error("cannot read the layout of the heads of the sessions", "store", h.store.Name(), "err", err)
		case ok && l.State != stateDone:
			// continues behind the cursor
			h.kick()
		case !ok:
			n, err := h.repo.Count()
			switch {
			case err != nil:
				slog.Error("cannot count the sessions", "store", h.store.Name(), "err", err)
			case n == 0:
				// nothing to build, and every session to come is written with its head
				if err := h.put(keyLayout, layout{Version: layoutVersion, State: stateDone}); err != nil {
					slog.Error("cannot write the layout of the heads of the sessions", "store", h.store.Name(), "err", err)
				}
			default:
				migrateWithoutHeads(h)
			}
		}
	}

	if bus, ok := core.FromContext[events.EventBus](h.ctx, ""); ok && bus != nil {
		unsubscribe := events.SubscribeFor(bus, func(backup.Restored) {
			if h.ctx.Err() == nil {
				h.kick()
			}
		})
		context.AfterFunc(h.ctx, unsubscribe)
	}
}

// repairMarks repairs the head of every session with a mark, reading only this session. A mark, which cannot be
// repaired, stays, and the searches stay right.
func (h *heads) repairMarks() {
	keys, err := h.keys(keyMarks)
	if err != nil {
		slog.Error("cannot list the open writes of the sessions", "store", h.store.Name(), "err", err)
		return
	}

	for _, key := range keys {
		if err := h.repair(ID(strings.TrimPrefix(key, keyMarks))); err != nil {
			slog.Error("cannot repair the head of a session with an open write", "session", strings.TrimPrefix(key, keyMarks), "err", err)
		}
	}
}

func (h *heads) repair(id ID) error {
	release := h.lock(id)
	defer release()

	var m mark
	ok, err := h.get(keyMarks+string(id), &m)
	if err != nil || !ok {
		return err
	}

	s, exists, err := h.load(id)
	if err != nil {
		return err
	}

	if exists {
		if err := h.settle(s); err != nil {
			return err
		}
	} else {
		for _, key := range headKeys(id, m.CreatedBy, m.ParentID) {
			if err := h.del(key); err != nil {
				return err
			}
		}
	}

	return h.del(keyMarks + string(id))
}

// kick starts a build in the background, unless one runs for the store, which then runs once more.
func (h *heads) kick() {
	key := shareKey(h.store)

	headShares.mu.Lock()
	defer headShares.mu.Unlock()

	sh := acquireShare(key)
	if sh.building {
		sh.again = true
		releaseShare(key, sh)
		return
	}

	// the reference is kept until the build has ended
	sh.building = true
	go h.run(key, sh)
}

func (h *heads) run(key any, sh *headShare) {
	for {
		err := blob.Derive(h.ctx, h.store, h.build)
		// a restore has ended the build, it starts from scratch once the restore has ended
		invalidated := errors.Is(err, blob.ErrInvalidated)
		if err != nil && !invalidated && h.ctx.Err() == nil {
			slog.Error("cannot build the heads of the sessions", "store", h.store.Name(), "err", err)
		}

		headShares.mu.Lock()
		if h.ctx.Err() == nil && (invalidated || sh.again) {
			sh.again = false
			headShares.mu.Unlock()
			continue
		}

		sh.building = false
		releaseShare(key, sh)
		headShares.mu.Unlock()
		return
	}
}

// headedRepository is the repository of the use cases with [Heads], which writes every session together with its head.
// Reads go to the sessions unchanged.
type headedRepository struct {
	Repository
	h *heads
}

func (r headedRepository) Save(s Session) error {
	return r.h.save(s)
}

func (r headedRepository) SaveAll(it iter.Seq[Session]) error {
	for s := range it {
		if err := r.Save(s); err != nil {
			return fmt.Errorf("cannot save %v: %w", s.ID, err)
		}
	}

	return nil
}

func (r headedRepository) DeleteByEntity(s Session) error {
	return r.h.remove(s)
}

func (r headedRepository) DeleteByID(id ID) error {
	opt, err := r.Repository.FindByID(id)
	switch {
	case err != nil && undecodable(err):
		// the heads cannot be located, but a head without its session does not count
		return r.h.remove(Session{ID: id})
	case err != nil:
		return err
	case opt.IsNone():
		return nil
	default:
		return r.h.remove(opt.Unwrap())
	}
}

func (r headedRepository) DeleteAllByID(ids iter.Seq[ID]) error {
	for id := range ids {
		if err := r.DeleteByID(id); err != nil {
			return err
		}
	}

	return nil
}

func (r headedRepository) DeleteAll() error {
	var ids []ID
	for id, err := range r.Repository.Identifiers() {
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	return r.DeleteAllByID(slices.Values(ids))
}

func (r headedRepository) Delete(predicate func(Session) (bool, error)) error {
	// collected first, because deleting while iterating the repository is not safe
	var victims []Session
	for s, err := range r.Repository.All() {
		if err != nil {
			return err
		}

		del, err := predicate(s)
		if errors.Is(err, data.SkipAll) {
			break
		}
		if err != nil {
			return err
		}

		if del {
			victims = append(victims, s)
		}
	}

	for _, s := range victims {
		if err := r.h.remove(s); err != nil {
			return err
		}
	}

	return nil
}
