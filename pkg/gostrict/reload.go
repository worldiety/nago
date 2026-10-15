// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"crypto/sha256"
	"encoding/hex"
	"go/types"
	"strings"
)

// A new version of a program, e.g. after an edit of a language model, may take over the state of the running one,
// so that an animation or an editor continues where it was. This is safe, if the state consists of the same data:
// the new version reads it only through its fields, which have the same names and types. A state, which holds
// handles, interfaces or functions, refers to objects of the former version, e.g. its three.js scene or its
// method tables, and is not taken over.

// stateHash returns the hash of the structure of a state, or "", if the state can not be taken over.
func stateHash(t types.Type) string {
	var sb strings.Builder
	if !writeShape(&sb, t, map[string]bool{}) {
		return ""
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(sum[:16])
}

// writeShape writes the structure of t, the names and the types of the fields of its structs, and reports false
// for a part, which belongs to an instance of a program.
func writeShape(sb *strings.Builder, t types.Type, seen map[string]bool) bool {
	t = types.Unalias(t)
	switch {
	case isHandle(t), isError(t):
		return false
	case isOpaque(t):
		n := t.(*types.Named)
		sb.WriteString(n.Obj().Pkg().Path() + "." + n.Obj().Name() + "[")
		for i := 0; i < n.TypeArgs().Len(); i++ {
			if i > 0 {
				sb.WriteByte(',')
			}
			if !writeShape(sb, n.TypeArgs().At(i), seen) {
				return false
			}
		}
		sb.WriteByte(']')
		return true
	}

	switch u := t.Underlying().(type) {
	case *types.Basic:
		sb.WriteString(typeKey(t))
		return true
	case *types.Pointer:
		sb.WriteByte('*')
		return writeShape(sb, u.Elem(), seen)
	case *types.Struct:
		key := typeKey(t)
		sb.WriteString(key)
		if seen[key] {
			return true
		}
		seen[key] = true
		sb.WriteByte('{')
		for f := range u.Fields() {
			sb.WriteString(f.Name() + " ")
			if !writeShape(sb, f.Type(), seen) {
				return false
			}
			sb.WriteByte(';')
		}
		sb.WriteByte('}')
		return true
	}
	// interfaces and functions belong to the instance
	return false
}

// CanTakeOver tells, whether the state of the type with the full name, which the former bundle created, may be
// passed to the functions of this bundle, see [Bundle.States].
func (b *Bundle) CanTakeOver(former *Bundle, state string) bool {
	h, ok := b.States[state]
	return ok && h != "" && former.States[state] == h
}
