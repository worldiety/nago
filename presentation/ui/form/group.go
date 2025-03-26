// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"reflect"
	"strings"
)

type Group struct {
	Name   string
	Fields []reflect.StructField
}

func GroupsFor[T any]() []Group {
	return GroupsOf(reflect.TypeFor[T]())
}

func GroupsOf(p reflect.Type, ignoreFields ...string) []Group {
	var res []Group
	//

	//typ := reflect.TypeOf(zero)
	//for i := 0; i < typ.NumField(); i++ {
	for _, field := range reflect.VisibleFields(p) {
		//field := typ.Field(i)

		if flag, ok := field.Tag.Lookup("visible"); ok && flag == "false" {
			continue
		}

		if !strings.HasPrefix(field.Name, "_") && !field.IsExported() {
			continue
		}

		ignored := false
		for _, ignoreField := range ignoreFields {
			if ignoreField == field.Name {
				ignored = true
				break
			}
		}

		if ignored {
			continue
		}

		sec := field.Tag.Get("section")
		var grp *Group
		for idx := range res {
			g := &res[idx]
			if g.Name == sec {
				grp = g
				break
			}
		}

		if grp == nil {
			res = append(res, Group{
				Name: sec,
			})

			grp = &res[len(res)-1]
		}

		grp.Fields = append(grp.Fields, field)
	}

	return res
}
