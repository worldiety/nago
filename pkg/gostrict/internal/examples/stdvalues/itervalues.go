// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/list"

// IterValues ranges over the iterator of List.Values, which a variable holds outside of the range.
func IterValues(a, b int32) int32 {
	l := list.Of(a, b, a+b)
	it := l.Values()
	sum := int32(0)
	for v := range it {
		sum = 10*sum + v
	}
	return sum
}
