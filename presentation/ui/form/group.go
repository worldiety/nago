// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/worldiety/i18n"
)

// Group represents a reflection-based field group (Group).
// Fields sharing the same "section" struct tag are collected under the same group.
type Group struct {
	Name   string
	Fields []reflect.StructField
}

func LocalizeGroups(bnd *i18n.Bundle, groups []Group) []Group {
	for idx := range groups {
		groups[idx].Name = bnd.Resolve(groups[idx].Name)
	}

	return groups
}

// GroupsFor returns field groups for the struct type parameter T.
// It uses the "section" tag to group fields.
func GroupsFor[T any]() []Group {
	return GroupsOf(reflect.TypeFor[T]())
}

// GroupsOf returns field groups for the given struct type.
// It groups fields by their "section" tag, skipping fields when:
// - tag `visible:"false"` is set
// - the field is unexported and does not start with "_" (private)
// - the field name appears in ignoreFields
func GroupsOf(p reflect.Type, ignoreFields ...string) []Group {
	var res []Group
	//

	if p.Kind() != reflect.Struct {
		panic(fmt.Errorf("type must be a struct but got %s", p.Kind()))
	}

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
