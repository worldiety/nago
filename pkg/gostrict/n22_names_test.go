// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import "go.wdy.de/nago/pkg/gostrict"

// The new names of the fast writer and reader (wokoda N22, L-N22-23, L-N22-27): Writer.Float64s and Writer.Int64s
// write a whole list, Reader.Float64s and Reader.Int64s read one, whose slice is valid until End. Only this file names
// them, so that the counter run against nago
// f4aca80c5, which does not know them, replaces it by testdata/n22/names_before.go.txt; that file writes and reads the
// lists element by element, as a host did before N22:
//
//	cd pkg/gostrict && go test -overlay testdata/n22/overlay.json -short -run TestN22 .

func n22WriteFloat64s(w gostrict.Writer, v []float64) { w.Float64s(v) }

func n22WriteInt64s(w gostrict.Writer, v []int64) { w.Int64s(v) }

func n22ReadFloat64s(r gostrict.Reader) []float64 { return r.Float64s() }

func n22ReadInt64s(r gostrict.Reader) []int64 { return r.Int64s() }
