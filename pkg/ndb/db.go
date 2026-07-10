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

// SeriesEngine is the capability of an [Engine] that stores typed, columnar
// time series (bucket-based columns keyed by unix-milli time). The tsdb engine
// implements it. The concrete, performance-critical typed read/write API lives
// on the engine's own package types; this neutral capability exposes only the
// engine-agnostic surface (naming the buckets/columns present). Callers that
// need the full API type-assert to the concrete engine and use its accessor.
type SeriesEngine interface {
	Engine

	// SeriesColumns lists the columns present in this engine instance as
	// "bucket/column" identifiers.
	SeriesColumns() ([]string, error)
}

// EngineKind identifies a storage engine implementation. It is the stable key
// under which an engine factory is registered (see [Register]) and that an
// engine instance reports via [Engine.Kind].
type EngineKind string

// EngineConfig is an engine-specific configuration value. It is passed through
// to the engine factory unchanged; the engine casts it to whatever type it
// understands.
//
// When a NEW engine instance is created it must be non-nil: a developer must
// consciously choose the engine settings rather than inherit implicit defaults.
// An explicit but empty value (e.g. an empty options struct, whose zero fields
// each engine documents and fills with sane internals) is accepted — it is the
// deliberate "I accept this engine's documented behaviour" gesture. A nil
// EngineConfig when creating an instance is an error.
//
// An engine may accept several shapes for convenience. The msgstore engine, for
// example, accepts a DSN string ("?compress=s2&split=64mib") or its native
// options struct for full programmatic control (custom strategy functions, a
// shared file pool). Which shapes are supported, and the DSN syntax, are
// documented by each engine.
type EngineConfig any

// Options configures how a [DB] (a directory of engine instances) is opened.
type Options struct {
	// FilePool is the bounded set of open file descriptors shared by every
	// engine instance opened by the DB. Sharing one pool across all engines
	// bounds the process-wide descriptor count regardless of how many engines
	// are active. If nil, a NewFilePool(1024) is created. The DB owns the
	// pool's lifecycle: it is closed by [DB.Close] and must never be closed by
	// an engine.
	FilePool *FilePool
}

// EngineOptions configures a single [DB.Engine] call.
type EngineOptions struct {
	// Kind selects the engine implementation. It is REQUIRED when creating a new
	// instance (there is no default engine kind). For an instance that already
	// exists on disk the recorded kind always wins, and Kind may be left empty;
	// if set, it must match the recorded kind.
	Kind EngineKind

	// Config is passed through to the engine factory unchanged (see
	// [EngineConfig]). It is REQUIRED (non-nil) when creating a new instance and
	// ignored when the instance already exists (its on-disk state is reused).
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
// pool is the DB's shared [FilePool]; every engine instance receives the same
// pool so the process holds one bounded set of file descriptors. Engines must
// route their file I/O through it and must never close it — its lifecycle is
// owned by the DB (see [DB.Close]).
//
// It returns the instance and a close function. The close function is retained
// solely by the owning [DB] and invoked by [DB.Close]; it is never exposed to
// consumers. This is why [Engine] itself has no Close: ownership stays with the
// DB, and a returned close func works across packages (unlike an unexported
// interface method, which a foreign package cannot implement).
type EngineFactory func(name, dir string, pool *FilePool, cfg EngineConfig) (eng Engine, close func() error, err error)

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

// lookupEngine resolves the factory for kind. kind must be non-empty (there is
// no default engine); an empty or unregistered kind is an error.
func lookupEngine(kind EngineKind) (EngineFactory, error) {
	if kind == "" {
		return nil, fmt.Errorf("ndb: no engine kind specified")
	}
	enginesMu.RLock()
	defer enginesMu.RUnlock()
	f, ok := engines[kind]
	if !ok {
		return nil, fmt.Errorf("ndb: no engine registered for kind %q", kind)
	}
	return f, nil
}
