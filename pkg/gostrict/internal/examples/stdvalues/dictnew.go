// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/dict"

// DictNew creates a map through dict.New as a value.
func DictNew(a, b int32) int32 {
	f := dict.New[string, int32]
	m := f()
	m.Set("a", a)
	m.Set("b", b)
	v := m.GetOr("b", 0)
	return 10*m.Len() + v
}
