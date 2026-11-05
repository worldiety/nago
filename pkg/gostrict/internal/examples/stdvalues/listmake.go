// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/list"

// ListMake creates a list through list.Make as a value.
func ListMake(a, b int32) int32 {
	f := list.Make[int32]
	l := f(a)
	l.Set(0, b)
	v, _ := l.Get(0)
	return 10*l.Len() + v
}
