// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"fmt"
	"go/types"

	"go.wdy.de/nago/pkg/gostrict/internal/vmcode"
)

// VM returns the program of the interpreter of package engine/vm, which the bundle carries next to its JavaScript,
// or the reason, why the program can not be compiled for it yet.
func (b *Bundle) VM() (*vmcode.Program, error) {
	if b.vm == nil && b.vmErr == nil {
		return nil, fmt.Errorf("the bundle has no program of the vm")
	}
	return b.vm, b.vmErr
}

// emitVM compiles the program for the interpreter of package engine/vm. The compiled program does not refer to the
// syntax and types, so the bundle does not keep them alive.
func (e *emitter) emitVM(b *Bundle, exports map[string]*types.Func) (p *vmcode.Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			p, err = nil, fmt.Errorf("vm: internal error: %v", r)
		}
	}()
	src := &vmcode.Source{
		Fset:      e.prog.Fset,
		Entries:   e.entries,
		Exports:   exports,
		Hosted:    map[string]bool{},
		Immutable: e.mut.immutable,
		Records:   map[string][]string{},
		Unions:    map[string]vmcode.Union{},
	}
	for _, p := range e.prog.Packages {
		src.Packages = append(src.Packages, &vmcode.Package{Path: p.Path, User: p.Kind == KindUser, Files: p.Files, Types: p.Types, Info: p.Info})
	}
	for _, x := range b.Externs {
		src.Hosted[x.Name] = true
	}
	src.HostVars = map[string]bool{}
	for _, p := range e.prog.Packages {
		if p.Kind != KindHeader {
			continue
		}
		for _, name := range p.Types.Scope().Names() {
			if v, ok := p.Types.Scope().Lookup(name).(*types.Var); ok && v.Exported() {
				src.HostVars[p.Path+"."+name] = true
			}
		}
	}
	for name, r := range b.Records {
		for _, f := range r.Fields {
			src.Records[name] = append(src.Records[name], f.Name)
		}
	}
	for name, u := range b.Unions {
		src.Unions[name] = vmcode.Union{Variants: u.Variants, Handles: u.Handles}
	}
	return vmcode.Compile(src)
}
