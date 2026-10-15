// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"testing"

	nui "go.wdy.de/nago/presentation/ui"
)

// TestEnumsB1 compares the constants of the enums of batch b1 with the constants of nago.
func TestEnumsB1(t *testing.T) {
	enums := []struct {
		name      string
		got, want int64
	}{
		{"TextFieldReduced", int64(TextFieldReduced), int64(nui.TextFieldReduced)},
		{"TextFieldOutlined", int64(TextFieldOutlined), int64(nui.TextFieldOutlined)},
		{"TextFieldBasic", int64(TextFieldBasic), int64(nui.TextFieldBasic)},
		{"TextFieldFilled", int64(TextFieldFilled), int64(nui.TextFieldFilled)},
		{"KeyboardDefault", int64(KeyboardDefault), int64(nui.KeyboardDefault)},
		{"KeyboardAscii", int64(KeyboardAscii), int64(nui.KeyboardAscii)},
		{"KeyboardInteger", int64(KeyboardInteger), int64(nui.KeyboardInteger)},
		{"KeyboardFloat", int64(KeyboardFloat), int64(nui.KeyboardFloat)},
		{"KeyboardEMail", int64(KeyboardEMail), int64(nui.KeyboardEMail)},
		{"KeyboardPhone", int64(KeyboardPhone), int64(nui.KeyboardPhone)},
		{"KeyboardSearch", int64(KeyboardSearch), int64(nui.KeyboardSearch)},
		{"KeyboardURL", int64(KeyboardURL), int64(nui.KeyboardURL)},
	}
	for _, e := range enums {
		if e.got != e.want {
			t.Errorf("%s = %d, want %d", e.name, e.got, e.want)
		}
	}
}
