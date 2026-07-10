// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import "errors"

var (
	errInvalidScheme    = errors.New("tsdb: invalid column scheme")
	errDecimalsTooLarge = errors.New("tsdb: decimals must be <= 18")
	errSchemeMismatch   = errors.New("tsdb: operation does not match column scheme")
	errClosed           = errors.New("tsdb: database is closed")
	errCorruptChunk     = errors.New("tsdb: corrupt chunk file")
	errCorruptBlock     = errors.New("tsdb: corrupt block")
	errBadName          = errors.New("tsdb: invalid bucket or column name")
	errColumnExists     = errors.New("tsdb: column already exists with a different scheme")
	errUnknownColumn    = errors.New("tsdb: unknown column")
)
