// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"slices"
	"testing"
	"unicode/utf8"
)

func TestNewPassword(t *testing.T) {
	var tmp []Password
	for range 100 {
		p := NewPassword()
		if !utf8.ValidString(string(p)) {
			t.Fatal("password is not valid utf8")
		}
		
		fmt.Println(p)

		tmp = append(tmp, p)
	}

	slices.Sort(tmp)
	tmp = slices.Compact(tmp)
	if len(tmp) != 100 {
		t.Fatal("length of tmp should be 10, generated duplicates")
	}
}
