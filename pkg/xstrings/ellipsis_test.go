// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xstrings

import "testing"

func TestEllipsisEnd(t *testing.T) {
	type args struct {
		s string
		n int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"0", args{"a", 1}, "a"},
		{"1", args{"abc", 1}, "..."},
		{"2", args{"Hallo Torben", 11}, "Hallo To..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EllipsisEnd(tt.args.s, tt.args.n); got != tt.want {
				t.Errorf("EllipsisEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}
