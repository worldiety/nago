// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"testing"

	"go.wdy.de/nago/pkg/xtime"
)

func Test_tsKey_Parse(t *testing.T) {
	type testCase struct {
		name    string
		s       tsKey
		want1   xtime.UnixMilliseconds
		wantErr bool
	}
	tests := []testCase{
		{
			name:    "empty",
			s:       tsKey(""),
			want1:   0,
			wantErr: true,
		},
		{
			name:  "valid",
			s:     tsKey("1767606558144"),
			want1: xtime.UnixMilliseconds(1767606558144),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, err := tt.s.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got1 != tt.want1 {
				t.Errorf("Parse() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
