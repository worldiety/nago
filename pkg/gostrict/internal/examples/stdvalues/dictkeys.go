// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/dict"

// DictKeys ranges over the iterator of the method value Map.Keys.
func DictKeys(a, b int32) int32 {
	m := dict.New[string, int32]()
	m.Set("a", a)
	m.Set("bb", b)
	f := m.Keys
	n := int32(0)
	for k := range f() {
		if k == "bb" {
			n += 10
		} else {
			n++
		}
	}
	return n + a
}
