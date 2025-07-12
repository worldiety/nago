// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"go.wdy.de/nago/presentation/ui"
	"testing"
	"time"
)

func Test_percentInYear(t *testing.T) {
	type args struct {
		year int
		t    time.Time
	}
	tests := []struct {
		name string
		args args
		want ui.Length
	}{
		{
			"start",
			args{2025, time.Date(2025, time.January, 1, 0, 0, 0, 0, time.Local)},
			"0%",
		},
		{
			"mid",
			args{2025, time.Date(2025, time.July, 2, 13, 0, 0, 0, time.Local)},
			"50%",
		},
		{
			"end",
			args{2025, time.Date(2025, time.December, 31, 24, 0, 0, -1, time.Local)},
			"100%",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cssPercentInYear(tt.args.year, tt.args.t); got != tt.want {
				t.Errorf("percentInYear() = %v, want %v", got, tt.want)
			}
		})
	}
}
