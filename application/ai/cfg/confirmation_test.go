// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import "testing"

func TestConfirmation(t *testing.T) {
	tests := []struct {
		name   string
		button Confirmation
		skip   bool
		marked bool // the operator confirms only marked tools
		want   bool
		// wantMarked is the expected confirmation of marked tools only
		wantMarked bool
	}{
		{name: "global default asks", button: ConfirmationGlobal, want: true},
		{name: "global skip", button: ConfirmationGlobal, skip: true, want: false},
		{name: "button always asks despite skip", button: ConfirmationAlways, skip: true, want: true},
		{name: "button never asks despite default", button: ConfirmationNever, want: false},
		{name: "marked does not ask for every change", button: ConfirmationMarked, want: false, wantMarked: true},
		{name: "marked despite skip", button: ConfirmationMarked, skip: true, wantMarked: true},
		{name: "operator confirms marked only", button: ConfirmationGlobal, marked: true, wantMarked: true},
		{name: "operator skip wins over marked only", button: ConfirmationGlobal, skip: true, marked: true},
		{name: "button always despite operator marked only", button: ConfirmationAlways, marked: true, want: true},
		{name: "button never despite operator marked only", button: ConfirmationNever, marked: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AssistantSettings{SkipConfirmation: tt.skip, ConfirmMarkedOnly: tt.marked}
			if got := tt.button.confirm(cfg); got != tt.want {
				t.Fatalf("want %v, got %v", tt.want, got)
			}

			if got := tt.button.confirmMarked(cfg); got != tt.wantMarked {
				t.Fatalf("marked: want %v, got %v", tt.wantMarked, got)
			}
		})
	}
}
