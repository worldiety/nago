// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"fmt"
	"testing"
)

func TestColor_WithChromaAndTone(t *testing.T) {

	tests := []struct {
		input   Color
		chroma  float64
		tone    float64
		want    Color
		wantErr bool
	}{
		{
			input:   "#1270E8FF",
			chroma:  66,
			tone:    49,
			want:    "#1270E8FF",
			wantErr: false,
		},
		{
			input:   "#1B8C30FF",
			chroma:  16,
			tone:    22,
			want:    "#293927FF",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt), func(t *testing.T) {
			got, err := tt.input.WithChromaAndTone(tt.chroma, tt.tone)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithChromaAndTone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WithChromaAndTone() got = %v, want %v", got, tt.want)
			}
		})
	}
}
