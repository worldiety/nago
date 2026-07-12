// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndb

import (
	"errors"

	"github.com/klauspost/compress/s2"
)

// ErrUnknownEncoding is returned by [Decompress] when a payload carries an
// [Encoding] marker that this build does not know how to decode. Because the
// marker is self-describing (see [Encoding]), an unknown value almost always
// signals either data corruption or a stored payload written by a newer engine.
var ErrUnknownEncoding = errors.New("ndb: unknown payload encoding")

// Decompress returns the verbatim payload bytes for a message, decoding the
// given [Encoding] if necessary. It is the engine-neutral counterpart to the
// compression an engine applies on write: a reader holds a self-describing
// [Message] and turns it back into raw bytes here without depending on any
// concrete engine package.
//
// For [EncodingRaw] the input slice is returned unchanged (no allocation, no
// copy) — callers that must retain it beyond the current iteration step still
// have to clone it, exactly as for a raw [Message.Payload]. For compressed
// encodings the result is a freshly allocated buffer of uncompressedLen bytes
// that the caller owns.
func Decompress(enc Encoding, payload []byte, uncompressedLen uint32) ([]byte, error) {
	switch enc {
	case EncodingRaw:
		return payload, nil
	case EncodingS2:
		return s2.Decode(make([]byte, uncompressedLen), payload)
	default:
		return nil, ErrUnknownEncoding
	}
}
