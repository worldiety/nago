// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/dict"

// DictIter ranges over the iterator of Map.All, which a variable holds outside of the range.
func DictIter(a, b int32) int32 {
	m := dict.New[string, int32]()
	m.Set("a", a)
	m.Set("b", b)
	it := m.All()
	sum := int32(0)
	for k, v := range it {
		if k == "b" {
			sum += 10 * v
		} else {
			sum += v
		}
	}
	return sum
}
