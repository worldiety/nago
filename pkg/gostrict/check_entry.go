// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/token"
	"go/types"
)

// checkEntries checks, that each exported function of a root of a whole program is an entry point: the host and
// the other packages call it within a frame of its own, see [Aspect], so it takes and returns only data. A value of
// a sealed interface would bring the code of its methods into the frame of the callee.
func checkEntries(prog *Program, d *diagnoser, decls map[*types.Func]funcSrc) {
	for _, p := range prog.Roots {
		if p.Kind != KindUser {
			pos := token.NoPos
			if len(p.Files) > 0 {
				pos = p.Files[0].Name.Pos()
			}
			d.report(pos, token.NoPos, CodeEntry, p.Path+" is no package of the subset", "the roots of a whole program are its scripts, which the subset checks", "pass only packages of the program as roots")
			continue
		}
		for _, name := range p.Types.Scope().Names() {
			fn, ok := p.Types.Scope().Lookup(name).(*types.Func)
			if !ok || !fn.Exported() {
				continue
			}
			src := decls[fn]
			pos, end := fn.Pos(), fn.Pos()+token.Pos(len(fn.Name()))
			if src.decl != nil {
				pos, end = src.decl.Type.Pos(), src.decl.Type.End()
			}
			const fix = "take and return records, lists, maps and basic values, or make the function unexported, if only its package uses it"
			if fn.Signature().TypeParams().Len() > 0 {
				d.report(pos, end, CodeEntry, fn.Name()+" is an entry point and can not be generic", "each exported function of a root package is an entry point, which the host calls by its name", fix)
				continue
			}
			b := newBoundary()
			sig, ok := b.exportSignature(name, fn.Signature())
			if !ok {
				d.report(pos, end, CodeEntry, "the entry point "+fn.Name()+" takes or returns values, which can not cross the boundary", "each exported function of a root package is an entry point, which the host calls with records, lists, maps, basic values and states", fix)
				continue
			}
			for _, t := range append(sig.Params, sig.Results...) {
				if hasUnion(t, b, map[string]bool{}) {
					d.report(pos, end, CodeEntry, "the entry point "+fn.Name()+" takes or returns a sealed interface", "an entry point runs in a frame with the rights of its package, where the methods of the values of the caller must not run", "pass a record with a kind field instead")
					break
				}
			}
		}
	}
}

// hasUnion tells, whether a value of t may hold a sealed interface of the program. A sealed interface of handles,
// e.g. a view of lib/ui, has no code of the program.
func hasUnion(t Type, b *boundary, seen map[string]bool) bool {
	switch t.Kind {
	case ValueUnion:
		return !b.unions[t.Name].Handles
	case ValueList, ValueOut:
		return hasUnion(*t.Elem, b, seen)
	case ValueMap:
		return hasUnion(*t.Key, b, seen) || hasUnion(*t.Elem, b, seen)
	case ValueRecord:
		if seen[t.Name] {
			return false
		}
		seen[t.Name] = true
		for _, f := range b.records[t.Name].Fields {
			if hasUnion(f.Type, b, seen) {
				return true
			}
		}
	}
	return false
}
