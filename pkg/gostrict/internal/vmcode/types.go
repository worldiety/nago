// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"go/token"
	"go/types"
)

// A value occupies slots in three banks: scalars (bool, int32, uint32 and float64 as bits), strings and
// references (objects, lists, maps, closures, handles, errors). A struct, whose fields are values, occupies the
// slots of its fields, so it lives in registers and is copied slot by slot without an allocation, which is exact,
// because the subset can not take the address of a variable or field. A large struct, which the program never
// changes, is a reference to a shared object instead, e.g. a style, which many cells of a table refer to.

// layout counts the slots of each bank, or is an offset into them.
type layout struct {
	n, s, r int32
}

func (a layout) add(b layout) layout {
	return layout{a.n + b.n, a.s + b.s, a.r + b.r}
}

func (a layout) max(b layout) layout {
	return layout{max(a.n, b.n), max(a.s, b.s), max(a.r, b.r)}
}

func (a layout) size() int32 {
	return a.n + a.s + a.r
}

type kind uint8

const (
	kInvalid kind = iota
	kBool
	kI32
	kU32
	kF64
	kStr
	// kStruct is a struct, whose fields occupy its slots
	kStruct
	// kShared is an immutable struct, whose value is a reference to an *object with its fields
	kShared
	// kPtr is a pointer to a struct, an *object with its fields
	kPtr
	kList
	kMap
	kOpt
	kOut
	kFunc
	// kIface is an interface: an *object with the type of the variant, or a Go error
	kIface
	kError
	// kHandle is a handle of a header package, whose slot holds the value of its js.Ref
	kHandle
	// kBig is the big.Int of std/big, whose slot holds a *math/big.Int
	kBig
	// kTuple are the results of a call
	kTuple
)

// tinfo describes a type of the program.
type tinfo struct {
	t    types.Type
	kind kind
	// l are the slots of a value
	l layout
	// fields are the fields of a struct, of the struct of a pointer or of a shared struct with their offsets in the
	// slots of the struct
	fields []field
	// obj are the slots of the object of a pointer or a shared struct
	obj layout
	// elem is the element of a list, the value of a map or option, the struct of a pointer
	elem *tinfo
	key  *tinfo
	// elems are the results of a tuple with their offsets
	elems []*tinfo
	offs  []layout
	// name is the full name of a named type, e.g. "example.com/app.Input"
	name string
	// methods are the methods of a variant of an interface by their name, filled lazily
	methods map[string]*Func
	sig     *types.Signature
	// zero is the zero value of a shared struct
	zero *object
	// rec are the indices of the fields of a record in the order of the boundary, see [compiler.plan]
	rec []int
	// handles is set for a union of handles, e.g. ui.View
	handles bool
}

type field struct {
	name string
	t    *tinfo
	off  layout
}

func (t *tinfo) String() string {
	return types.TypeString(t.t, nil)
}

// slotKind returns the bank of a single slot type, or false.
func (t *tinfo) single() (bank uint8, ok bool) {
	switch {
	case t.l == layout{n: 1}:
		return bankN, true
	case t.l == layout{s: 1}:
		return bankS, true
	case t.l == layout{r: 1}:
		return bankR, true
	}
	return 0, false
}

const (
	bankN uint8 = iota
	bankS
	bankR
)

// maxFlat is the number of slots, up to which an immutable struct lives in slots instead of a shared object.
const maxFlat = 8

const (
	stdPrefix = "go.wdy.de/nago/pkg/gostrict/std/"
	listPath  = stdPrefix + "list"
	dictPath  = stdPrefix + "dict"
	optPath   = stdPrefix + "opt"
	msgPath   = stdPrefix + "msg"
	jsPath    = stdPrefix + "js"
	bigPath   = stdPrefix + "big"
	textPath  = stdPrefix + "text"
	numPath   = stdPrefix + "num"
)

// typeKey is the full name of a type, e.g. "example.com/app.Input".
func typeKey(t types.Type) string {
	return types.TypeString(t, func(p *types.Package) string { return p.Path() })
}

var errorType = types.Universe.Lookup("error").Type()

// typeOf returns the description of t, which contains no type parameters.
func (c *compiler) typeOf(t types.Type) *tinfo {
	key := typeKey(t)
	if ti, ok := c.types[key]; ok {
		return ti
	}
	ti := &tinfo{t: t}
	// recursive types refer to themselves through pointers, lists and maps
	c.types[key] = ti
	c.describe(ti)
	return ti
}

func (c *compiler) describe(ti *tinfo) {
	t := ti.t
	if n, ok := types.Unalias(t).(*types.Named); ok {
		ti.name = typeKey(t)
		if n.Obj().Pkg() != nil {
			p, name := n.Obj().Pkg().Path(), n.Obj().Name()
			targs := n.TypeArgs()
			switch {
			case p == listPath && (name == "List" || name == "Frozen"):
				ti.kind, ti.l = kList, layout{r: 1}
				ti.elem = c.typeOf(targs.At(0))
				return
			case p == msgPath && name == "Out":
				ti.kind, ti.l = kOut, layout{r: 1}
				ti.elem = c.typeOf(targs.At(0))
				return
			case p == dictPath && name == "Map":
				ti.kind, ti.l = kMap, layout{r: 1}
				ti.key, ti.elem = c.typeOf(targs.At(0)), c.typeOf(targs.At(1))
				return
			case p == optPath && name == "Option":
				ti.kind, ti.l = kOpt, layout{r: 1}
				ti.elem = c.typeOf(targs.At(0))
				return
			case p == bigPath && name == "Int":
				ti.kind, ti.l = kBig, layout{r: 1}
				return
			}
			if isHandle(t) {
				ti.kind, ti.l = kHandle, layout{r: 1}
				return
			}
		}
	}
	if types.Identical(t, errorType) {
		ti.kind, ti.l = kError, layout{r: 1}
		return
	}

	switch u := t.Underlying().(type) {
	case *types.Basic:
		switch u.Kind() {
		case types.Bool, types.UntypedBool:
			ti.kind, ti.l = kBool, layout{n: 1}
		case types.Int32, types.UntypedRune:
			ti.kind, ti.l = kI32, layout{n: 1}
		case types.Uint32:
			ti.kind, ti.l = kU32, layout{n: 1}
		case types.Float64, types.UntypedFloat:
			ti.kind, ti.l = kF64, layout{n: 1}
		case types.String, types.UntypedString:
			ti.kind, ti.l = kStr, layout{s: 1}
		default:
			c.fail(token.NoPos, "type %s is not supported yet", t)
		}
	case *types.Struct:
		var off layout
		for i := range u.NumFields() {
			f := u.Field(i)
			ft := c.typeOf(f.Type())
			ti.fields = append(ti.fields, field{name: f.Name(), t: ft, off: off})
			off = off.add(ft.l)
		}
		ti.obj = off
		ti.kind, ti.l = kStruct, off
		if off.size() > maxFlat && c.src.Immutable(t) {
			ti.kind, ti.l = kShared, layout{r: 1}
		}
	case *types.Pointer:
		// the kind and slots come first, the struct may refer to its pointer, e.g. type Node struct{ Next *Node }
		ti.kind, ti.l = kPtr, layout{r: 1}
		elem := c.typeOf(u.Elem())
		ti.elem = elem
		ti.fields, ti.obj = elem.fields, elem.obj
		if elem.fields == nil && elem.obj == (layout{}) {
			// the struct is still being described, e.g. type Node struct{ Next *Node }
			c.later = append(c.later, func() { ti.fields, ti.obj = elem.fields, elem.obj })
		}
	case *types.Signature:
		ti.kind, ti.l, ti.sig = kFunc, layout{r: 1}, u
	case *types.Interface:
		ti.kind, ti.l = kIface, layout{r: 1}
	case *types.Tuple:
		panic("vm: use tupleOf")
	default:
		c.fail(token.NoPos, "type %s is not supported yet", t)
	}
}

// tupleOf describes the results of a signature.
func (c *compiler) tupleOf(res *types.Tuple) *tinfo {
	if res.Len() == 1 {
		return c.typeOf(res.At(0).Type())
	}
	ti := &tinfo{kind: kTuple}
	for v := range res.Variables() {
		et := c.typeOf(v.Type())
		ti.elems = append(ti.elems, et)
		ti.offs = append(ti.offs, ti.l)
		ti.l = ti.l.add(et.l)
	}
	return ti
}

// tuple describes a tuple of the given types.
func tuple(ts ...*tinfo) *tinfo {
	if len(ts) == 1 {
		return ts[0]
	}
	ti := &tinfo{kind: kTuple}
	for _, et := range ts {
		ti.elems = append(ti.elems, et)
		ti.offs = append(ti.offs, ti.l)
		ti.l = ti.l.add(et.l)
	}
	return ti
}

// isHandle tells, whether t is a handle of a header package: a named struct with a single unexported field of
// type js.Ref.
func isHandle(t types.Type) bool {
	n, ok := types.Unalias(t).(*types.Named)
	if !ok || n.Obj().Pkg() == nil {
		return false
	}
	st, ok := n.Underlying().(*types.Struct)
	if !ok || st.NumFields() != 1 || st.Field(0).Exported() {
		return false
	}
	ft, ok := types.Unalias(st.Field(0).Type()).(*types.Named)
	return ok && ft.Obj().Pkg() != nil && ft.Obj().Pkg().Path() == jsPath && ft.Obj().Name() == "Ref"
}

// fieldOf returns the field of a struct type by name.
func (t *tinfo) fieldOf(name string) field {
	for _, f := range t.fields {
		if f.name == name {
			return f
		}
	}
	panic("vm: no field " + name + " in " + t.String())
}
