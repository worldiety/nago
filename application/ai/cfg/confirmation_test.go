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
		want   bool
	}{
		{name: "global default asks", button: ConfirmationGlobal, want: true},
		{name: "global skip", button: ConfirmationGlobal, skip: true, want: false},
		{name: "button always asks despite skip", button: ConfirmationAlways, skip: true, want: true},
		{name: "button never asks despite default", button: ConfirmationNever, want: false},
		{name: "marked does not ask for every change", button: ConfirmationMarked, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.button.confirm(AssistantSettings{SkipConfirmation: tt.skip}); got != tt.want {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}
