// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndb

import (
	"fmt"
	"iter"
	"maps"
	"slices"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/blob"
)

// Store is the blob key/value store contract, re-exported from the blob package
// so that ndb consumers have a single import for both message and blob storage.
type Store = blob.Store

// Engine is a single, technically distinct nago storage engine instance managed
// by a [DB]. Each engine has its own physical folder and a unique name and is
// backed by exactly one engine implementation (see [EngineKind]), e.g. the
// msgstore event engine or a blob-backed engine. A DB may hold many engine
// instances, including several of the same kind (e.g. two independent msgstore
// directories "audit" and "telemetry").
//
// Engine itself is intentionally minimal. The container-level capabilities an
// instance provides depend on its kind and are obtained via a type assertion to
// the matching capability interface ([BlobEngine], [MessageEngine], ...). This
// keeps the storage layer free to grow and to mix engine generations within one
// process without forcing every engine to implement every feature, and without
// option-wrapper boilerplate on the hot path:
//
//	if me, ok := eng.(ndb.MessageEngine); ok {
//		me.Messages().Append(typeID, trace, payload)
//	}
//
// Engine deliberately does NOT expose Close: the lifecycle of every engine is
// owned by the [DB] that created it (see [DB.Close]). This makes it impossible
// for a consumer to close an engine out from under the DB's cache and hand the
// next caller a dead instance.
type Engine interface {
	// Name returns the unique engine instance name. Two engines with the same
	// name are the same logical instance.
	Name() string

	// Kind returns the identifier of the engine implementation backing this
	// instance (see [EngineKind]). It lets callers reason about which on-disk
	// format and feature set to expect.
	Kind() EngineKind
}

// BlobEngine is the capability of an [Engine] that manages a set of named blob
// stores. It is the successor of the blob.Stores meta-interface; existing
// blob backends (tdb, fs, mem) are intended to migrate behind this capability.
type BlobEngine interface {
	Engine

	// Store opens or creates the named blob store within this engine instance.
	Store(name string, opts blob.OpenStoreOptions) (blob.Store, error)

	// LookupStore returns an existing named blob store, never creating one.
	LookupStore(name string) (option.Opt[blob.Store], error)

	// Stores lists the blob stores present in this engine instance.
	Stores() iter.Seq2[blob.StoreInfo, error]
}

// MessageEngine is the capability of an [Engine] that provides a durable
// message log (see [Messages]). The msgstore engine implements it.
type MessageEngine interface {
	Engine

	// Messages returns the message capability of this engine instance.
	Messages() Messages
}

// EngineKind identifies a storage engine implementation. It is the stable key
// under which an engine factory is registered (see [Register]) and that an
// engine instance reports via [Engine.Kind].
type EngineKind string

// EngineConfig is an engine-specific configuration value. It is passed through
// to the engine factory unchanged; the engine casts it to whatever type it
// understands. A nil EngineConfig means "engine defaults".
//
// An engine may accept several shapes for convenience. The msgstore engine, for
// example, accepts nil (defaults), a DSN string ("?compress=s2&split=64mib"),
// or its native options struct for full programmatic control (custom strategy
// functions, a shared file pool). Which shapes are supported, and the DSN
// syntax, are documented by each engine.
type EngineConfig any

// Options configures how a [DB] (a directory of engine instances) is opened.
type Options struct {
	// DefaultKind selects the engine implementation used for instances that do
	// not yet exist on disk and must be created, whenever [EngineOptions.Kind]
	// is empty. Existing instances are always reopened with the engine that
	// created them (recorded in a per-instance marker file), regardless of this
	// setting, so that an on-disk format is never misinterpreted. If empty, a
	// single registered engine is used when exactly one exists; otherwise
	// opening a missing instance is an error.
	DefaultKind EngineKind
}

// EngineOptions configures a single [DB.Engine] call.
type EngineOptions struct {
	// Kind selects the engine implementation for a NEW instance. For an
	// existing instance the recorded kind always wins. If empty, the DB's
	// [Options.DefaultKind] (then the sole registered engine) is used.
	Kind EngineKind

	// Config is passed through to the engine factory unchanged (see
	// [EngineConfig]). It is ignored when the instance is already open.
	Config EngineConfig
}

// EngineInfo describes one engine instance present in a [DB].
type EngineInfo struct {
	Name string
	Kind EngineKind
}

// EngineFactory opens (or creates) a single engine instance rooted at dir with
// the given name and engine-specific config. Engine implementations register a
// factory via [Register] so that ndb can instantiate engines without depending
// on any concrete engine package (which would otherwise create an import cycle,
// since engines depend on this contract package).
//
// It returns the instance and a close function. The close function is retained
// solely by the owning [DB] and invoked by [DB.Close]; it is never exposed to
// consumers. This is why [Engine] itself has no Close: ownership stays with the
// DB, and a returned close func works across packages (unlike an unexported
// interface method, which a foreign package cannot implement).
type EngineFactory func(name, dir string, cfg EngineConfig) (eng Engine, close func() error, err error)

var (
	enginesMu sync.RWMutex
	engines   = map[EngineKind]EngineFactory{}
)

// Register makes an engine implementation available to [Open] under kind.
// It is intended to be called from an engine package's init function. It panics
// if kind is empty, factory is nil, or kind is already registered, because
// these are programming errors that must surface at startup.
func Register(kind EngineKind, factory EngineFactory) {
	if kind == "" {
		panic("ndb: cannot register engine with empty kind")
	}
	if factory == nil {
		panic("ndb: cannot register nil engine factory for " + string(kind))
	}

	enginesMu.Lock()
	defer enginesMu.Unlock()
	if _, exists := engines[kind]; exists {
		panic("ndb: engine already registered: " + string(kind))
	}
	engines[kind] = factory
}

// RegisteredEngines returns the kinds of all currently registered engines,
// sorted for stable output.
func RegisteredEngines() []EngineKind {
	enginesMu.RLock()
	defer enginesMu.RUnlock()
	kinds := slices.Collect(maps.Keys(engines))
	slices.Sort(kinds)
	return kinds
}

// lookupEngine resolves the factory for kind. When kind is empty it falls back
// to the sole registered engine. It also returns the resolved kind so that the
// caller can record it (e.g. in the per-instance marker) rather than an empty
// string.
func lookupEngine(kind EngineKind) (EngineFactory, EngineKind, error) {
	enginesMu.RLock()
	defer enginesMu.RUnlock()

	if kind != "" {
		f, ok := engines[kind]
		if !ok {
			return nil, "", fmt.Errorf("ndb: no engine registered for kind %q", kind)
		}
		return f, kind, nil
	}

	// no explicit kind: fall back to the sole registered engine, if unambiguous
	switch len(engines) {
	case 0:
		return nil, "", fmt.Errorf("ndb: no storage engines registered")
	case 1:
		for k, f := range engines {
			return f, k, nil
		}
	}
	return nil, "", fmt.Errorf("ndb: multiple engines registered, set Options.DefaultKind or EngineOptions.Kind")
}
