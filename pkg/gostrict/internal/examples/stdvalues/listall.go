// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/list"

// ListAll ranges over the iterator of the method value List.All.
func ListAll(a, b int32) int32 {
	l := list.Of(a, b, a+b)
	f := l.All
	sum := int32(0)
	for i, v := range f() {
		sum += (i + 1) * v
	}
	return sum
}
