// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/list"

// IterAll ranges over the iterator of List.All, which a variable holds outside of the range.
func IterAll(a, b int32) int32 {
	l := list.Of(a, b, a+b)
	it := l.All()
	sum := int32(0)
	for i, v := range it {
		sum += (i + 1) * v
	}
	return sum
}
