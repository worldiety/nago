// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package concurrent

import "testing"

func TestCompareAndSwap(t *testing.T) {
	var destroyed Value[bool]
	if !CompareAndSwap(&destroyed, false, true) {
		t.Fatal("unreachable")
	}

	if CompareAndSwap(&destroyed, false, true) {
		t.Fatal("unreachable")
	}
}
