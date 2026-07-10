package msgstore

import (
	"fmt"

	"go.wdy.de/nago/pkg/ndb"
)

// EngineKind is the [ndb.EngineKind] under which this event engine registers
// itself. Engine instances created by this engine record it in their marker
// file so they are always reopened with the matching on-disk format.
const EngineKind ndb.EngineKind = "msgstore"

func init() {
	ndb.Register(EngineKind, openEngine)
}

// engine adapts a *DB to the [ndb.Engine] / [ndb.MessageEngine] contracts.
// msgstore provides a durable message log but no blob stores, so it implements
// MessageEngine but not BlobEngine.
type engine struct {
	name string
	db   *DB
}

var (
	_ ndb.Engine        = (*engine)(nil)
	_ ndb.MessageEngine = (*engine)(nil)
)

// openEngine is the [ndb.EngineFactory] for the msgstore engine. ndb calls it to
// open or create the engine instance rooted at dir.
//
// cfg accepts three shapes:
//
//   - nil: engine defaults are used.
//   - string: a DSN such as "?compress=s2&split=64mib&maxmsg=16mib" (see
//     [parseDSN]). Convenient and declarative, but limited to scalar settings.   - [Options]: the native options struct for full programmatic control,
//     including custom Compress/ShouldSplit functions and a shared FilePool.
func openEngine(name, dir string, pool *ndb.FilePool, cfg ndb.EngineConfig) (ndb.Engine, func() error, error) {
	var opts Options
	switch c := cfg.(type) {
	case nil:
		// zero Options; resolve() fills in defaults
	case string:
		o, err := parseDSN(c)
		if err != nil {
			return nil, nil, fmt.Errorf("msgstore: invalid config DSN: %w", err)
		}
		opts = o
	case Options:
		opts = c
	default:
		return nil, nil, fmt.Errorf("msgstore: unsupported engine config type %T", cfg)
	}

	// Use the shared pool injected by ndb unless the caller explicitly supplied
	// one (via Options.FilePool or the filepool=N DSN option), which takes
	// precedence. This keeps all engines of a DB on a single bounded pool while
	// still allowing an explicit override.
	if opts.FilePool == nil {
		opts.FilePool = pool
	}

	db, err := Open(dir, opts)
	if err != nil {
		return nil, nil, err
	}
	return &engine{name: name, db: db}, db.Close, nil
}

func (e *engine) Name() string { return e.name }

func (e *engine) Kind() ndb.EngineKind { return EngineKind }

func (e *engine) Messages() ndb.Messages { return e.db }

// DB exposes the underlying engine handle for callers that need msgstore-specific
// features beyond the neutral [ndb.Messages] contract (such as RebuildTimeIndex).
//
// Do not Close the returned *DB yourself: its lifecycle is owned by the
// [ndb.DB] that opened this instance. This accessor is deliberately a separate,
// non-idiomatic step so it is not reached for by accident.
func (e *engine) DB() *DB { return e.db }
