// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/list"

// ListNew creates a list through list.New as a value.
func ListNew(a, b int32) int32 {
	f := list.New[int32]
	l := f()
	l.Append(a, b)
	v, _ := l.Get(1)
	return 10*l.Len() + v
}
