package msgstore

// Encoding describes the compression algorithm used for a message payload.
type Encoding uint8

const (
	// EncodingRaw means no compression, payload is stored as-is.
	EncodingRaw Encoding = 0
	// EncodingS2 means S2 compression (Snappy-compatible, faster).
	EncodingS2 Encoding = 1
)

// TypeID identifies an event type. It corresponds to the numeric directory name under events/.
type TypeID uint64

