// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"fmt"
	"go/token"
	"go/types"
	"strings"

	jsoncore "go.wdy.de/nago/pkg/gostrict/std/json/core"
)

// jsonType requests the description of t, which $json_Decode and $json_Encode of the runtime read, and returns the name of the
// function, which returns it. The function creates the description once and refers to the descriptions of the elements
// and fields by their functions, so that a recursive type, e.g. a struct with a list of itself, needs no order. The
// checker allows only the types, which it describes, see [CodeJSON].
func (e *emitter) jsonType(t types.Type) string {
	key := "J:" + typeKey(t)
	if n, ok := e.names[key]; ok {
		return n
	}
	name := e.ident("$J$" + typeBase(t))
	e.names[key] = name

	desc := ""
	if n, ok := types.Unalias(t).(*types.Named); ok && n.Obj().Pkg() != nil {
		targs := n.TypeArgs()
		switch n.Obj().Pkg().Path() + "." + n.Obj().Name() {
		case listPath + ".List", listPath + ".Frozen":
			desc = `{ k: "l", e: ` + e.jsonType(targs.At(0)) + ` }`
		case dictPath + ".Map":
			desc = `{ k: "m", e: ` + e.jsonType(targs.At(1)) + ` }`
		case optPath + ".Option":
			desc = `{ k: "o", e: ` + e.jsonType(targs.At(0)) + ` }`
		case binPath + ".Bytes":
			desc = `{ k: "y" }`
		case bigPath + ".Int":
			desc = `{ k: "n" }`
		case decimalPath + ".Dec":
			// a decimal number, whose integer and scale are the fields M and S of the record
			st := t.Underlying().(*types.Struct)
			props := map[string]string{}
			for i := range st.NumFields() {
				props[st.Field(i).Name()] = fieldName(st.Field(i), i)
			}
			desc = `{ k: "d", z: () => ` + e.resultZero(t) + `, m: ` + jsString(props["M"]) + `, s: ` + jsString(props["S"]) + ` }`
		}
	}
	if desc == "" && isStruct(t) {
		st := t.Underlying().(*types.Struct)
		var fields []string
		seen := map[string]bool{}
		for i := range st.NumFields() {
			f := st.Field(i)
			if !f.Exported() {
				continue
			}
			k, ok := jsoncore.Key(f.Name(), st.Tag(i))
			if !ok || seen[k] {
				continue
			}
			seen[k] = true
			omit := "false"
			if jsoncore.OmitEmpty(st.Tag(i)) {
				omit = "true"
			}
			fields = append(fields, "["+jsString(k)+", ["+jsString(fieldName(f, i))+", "+e.jsonType(f.Type())+", "+omit+"]]")
		}
		desc = `{ k: "r", z: () => ` + e.resultZero(t) + `, f: new Map([` + strings.Join(fields, ", ") + `]) }`
	}
	if desc == "" {
		kinds := map[types.BasicKind]string{types.Bool: "b", types.Int32: "i", types.Uint32: "u", types.Int64: "I", types.Uint64: "U", types.Float64: "f", types.String: "s"}
		k, ok := kinds[basicKind(t)]
		if !ok {
			fail(token.NoPos, "json.Decode can not read %s", t)
		}
		desc = `{ k: "` + k + `" }`
	}

	fmt.Fprintf(&e.out, "function %s() { return %s.d || (%s.d = %s); }\n", name, name, name, desc)
	return name
}
