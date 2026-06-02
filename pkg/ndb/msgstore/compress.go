package msgstore

import (
	"github.com/klauspost/compress/s2"
)

// CompressFunc decides per message how to compress the payload.
// The typeID is the primary decision parameter (e.g. "type 42 always compress").
// The payload is available as secondary criterion (e.g. skip compression for tiny payloads).
// Returns the encoding marker and the (possibly compressed) output bytes.
type CompressFunc func(typeID TypeID, payload []byte) (Encoding, []byte)

// NoCompression returns the payload unchanged with EncodingRaw.
func NoCompression(_ TypeID, payload []byte) (Encoding, []byte) {
	return EncodingRaw, payload
}

// AlwaysS2 always compresses the payload with S2.
func AlwaysS2(_ TypeID, payload []byte) (Encoding, []byte) {
	return EncodingS2, s2.Encode(nil, payload)
}

// DefaultCompression compresses with S2 when len(payload) > 512 and the
// compressed output is actually smaller. Otherwise returns raw.
func DefaultCompression(_ TypeID, payload []byte) (Encoding, []byte) {
	if len(payload) <= 512 {
		return EncodingRaw, payload
	}
	compressed := s2.Encode(nil, payload)
	if len(compressed) >= len(payload) {
		return EncodingRaw, payload
	}
	return EncodingS2, compressed
}

// Decompress decompresses a message payload according to its encoding.
func Decompress(enc Encoding, compressed []byte, uncompressedLen uint32) ([]byte, error) {
	switch enc {
	case EncodingRaw:
		return compressed, nil
	case EncodingS2:
		dst := make([]byte, uncompressedLen)
		return s2.Decode(dst, compressed)
	default:
		return nil, ErrCorruptCRC // unknown encoding treated as corruption
	}
}

