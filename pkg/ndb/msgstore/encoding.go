package msgstore

import "go.wdy.de/nago/pkg/ndb"

// Encoding describes the compression algorithm used for a message payload.
// It is an alias of [ndb.Encoding] so that engine-internal code stays terse
// while remaining byte-for-byte identical to the neutral contract type. Using
// an alias (not a fresh type) is what lets *DB satisfy the ndb interfaces with
// zero conversion on the hot path.
type Encoding = ndb.Encoding

const (
	// EncodingRaw means no compression, payload is stored as-is.
	EncodingRaw = ndb.EncodingRaw
	// EncodingS2 means S2 compression (Snappy-compatible, faster).
	EncodingS2 = ndb.EncodingS2
)

// TypeID identifies an event type. It corresponds to the numeric directory name
// under events/. Alias of [ndb.TypeID]; see [Encoding] for the rationale.
type TypeID = ndb.TypeID

// Seq is the global strict-monotonic sequence number. Alias of [ndb.Seq].
type Seq = ndb.Seq
