// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/dict"

// DictGet reads a map through the method value Get.
func DictGet(a, b int32) int32 {
	m := dict.New[string, int32]()
	m.Set("a", a)
	f := m.Get
	v, ok := f("a")
	if !ok {
		return -1
	}
	return v + b
}
