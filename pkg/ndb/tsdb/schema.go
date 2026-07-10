// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

// Scheme identifies the exact storage scheme of a column. Every column has
// exactly one scheme fixed at creation time. The scheme determines how values
// are encoded on disk and which typed read/write API is valid for the column.
type Scheme uint8

const (
	// SchemeInvalid is the zero value and never valid.
	SchemeInvalid Scheme = iota

	// SchemeDecimal stores fixed-decimal numeric values: the caller-supplied
	// float64 (with noise) is rounded to a scaled int64 using the column's
	// Decimals, then delta+zig-zag encoded. Equidistant or mostly-equidistant
	// unix-millis keys with holes. Read via I64 (raw scaled) or F64 (unscaled).
	SchemeDecimal

	// SchemeEnum stores a string enum via a dictionary. The value stream holds
	// varint dictionary ids; the dictionary is append-only and extends at
	// runtime. Read via String (or the raw enum id).
	SchemeEnum

	// SchemeString stores an arbitrary variable string per timestamp. Values
	// are length-prefixed and the block is block-compressed. Read via String.
	SchemeString
)

func (s Scheme) String() string {
	switch s {
	case SchemeDecimal:
		return "decimal"
	case SchemeEnum:
		return "enum"
	case SchemeString:
		return "string"
	default:
		return "invalid"
	}
}

func (s Scheme) valid() bool {
	return s == SchemeDecimal || s == SchemeEnum || s == SchemeString
}

// numeric reports whether the scheme carries int64/float64 values (as opposed
// to strings).
func (s Scheme) numeric() bool {
	return s == SchemeDecimal
}

// Schema is the persisted, immutable-after-creation description of a column.
// It is stored as schema.json in the column directory and written atomically
// (temp file + rename).
type Schema struct {
	// Scheme is the column storage scheme.
	Scheme Scheme `json:"scheme"`

	// Decimals is the number of significant decimal places used to scale
	// float64 inputs to int64 for SchemeDecimal. Ignored for other schemes.
	// The developer chooses this to truncate float noise.
	Decimals uint8 `json:"decimals,omitempty"`

	// Version of the on-disk format for this column.
	Version uint8 `json:"version"`
}

const schemaVersion uint8 = 1

func (s Schema) validate() error {
	if !s.Scheme.valid() {
		return errInvalidScheme
	}
	if s.Decimals > 18 {
		return errDecimalsTooLarge
	}
	return nil
}
