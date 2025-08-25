// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"encoding/json"
	"go.wdy.de/nago/application/permission"
	"reflect"
	"testing"
)

func TestResource_MarshalJSON(t *testing.T) {
	tmp := map[Resource][]permission.ID{}
	tmp[Resource{
		Name: "a",
		ID:   "b",
	}] = []permission.ID{"1", "2", "3"}

	buf, err := json.Marshal(tmp)
	if err != nil {
		t.Fatal(err)
	}

	var tmp2 map[Resource][]permission.ID
	err = json.Unmarshal(buf, &tmp2)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(tmp, tmp2) {
		t.Errorf("%+v != %+v", tmp, tmp2)
	}
}

func TestEmail_Valid(t *testing.T) {
	tests := []struct {
		name string
		e    Email
		want bool
	}{
		{
			name: ".group",
			e:    "vorname.nachname@example.group",
			want: true,
		},

		{
			name: ".allfinanzberatung",
			e:    "vorname.nachname@example.allfinanzberatung",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Valid(); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
