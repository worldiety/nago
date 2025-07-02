// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package semver

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name  string
		args  args
		want  Version
		want1 bool
	}{
		{
			name:  "test1",
			args:  args{str: "1.2.3"},
			want:  Version{Major: 1, Minor: 2, Patch: 3},
			want1: true,
		},
		{
			name:  "test2",
			args:  args{str: "v1.2.3"},
			want:  Version{Major: 1, Minor: 2, Patch: 3},
			want1: true,
		},
		{
			name:  "test3",
			args:  args{str: "v1"},
			want:  Version{Major: 1, Minor: 0, Patch: 0},
			want1: true,
		},
		{
			name:  "test4",
			args:  args{str: "v1.2"},
			want:  Version{Major: 1, Minor: 2, Patch: 0},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := Parse(tt.args.str)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Parse() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestVersion_Newer(t *testing.T) {
	type fields struct {
		Major int
		Minor int
		Patch int
	}
	type args struct {
		other Version
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "test1",
			fields: fields{Major: 1, Minor: 2, Patch: 3},
			args:   args{other: Version{Major: 1, Minor: 2, Patch: 3}},
			want:   false,
		},
		{
			name:   "test2",
			fields: fields{Major: 0, Minor: 2, Patch: 3},
			args:   args{other: Version{Major: 1, Minor: 2, Patch: 3}},
			want:   false,
		},
		{
			name:   "test3",
			fields: fields{Major: 2, Minor: 2, Patch: 3},
			args:   args{other: Version{Major: 1, Minor: 2, Patch: 3}},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Patch: tt.fields.Patch,
			}
			if got := v.Newer(tt.args.other); got != tt.want {
				t.Errorf("Newer() = %v, want %v", got, tt.want)
			}
		})
	}
}
