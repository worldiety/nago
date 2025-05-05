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
