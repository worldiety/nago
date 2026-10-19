// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"fmt"
	"strings"
)

// opNames are the names of the operations, e.g. ldkn for opLdkN.
var opNames = [...]string{
	opNop:        "nop",
	opMovN:       "movn",
	opMovS:       "movs",
	opMovR:       "movr",
	opLdiN:       "ldin",
	opLdkN:       "ldkn",
	opLdkS:       "ldks",
	opLdkR:       "ldkr",
	opLdgN:       "ldgn",
	opLdgS:       "ldgs",
	opLdgR:       "ldgr",
	opStgN:       "stgn",
	opStgS:       "stgs",
	opStgR:       "stgr",
	opLdoN:       "ldon",
	opLdoS:       "ldos",
	opLdoR:       "ldor",
	opStoN:       "ston",
	opStoS:       "stos",
	opStoR:       "stor",
	opNew:        "new",
	opAddI:       "addi",
	opAddIK:      "addik",
	opSubI:       "subi",
	opMulI:       "muli",
	opDivI:       "divi",
	opRemI:       "remi",
	opDivU:       "divu",
	opRemU:       "remu",
	opAndI:       "andi",
	opOrI:        "ori",
	opXorI:       "xori",
	opAndNotI:    "andnoti",
	opShlI:       "shli",
	opShrI:       "shri",
	opShrU:       "shru",
	opNegI:       "negi",
	opNotI:       "noti",
	opEqI:        "eqi",
	opNeI:        "nei",
	opLtI:        "lti",
	opLeI:        "lei",
	opLtU:        "ltu",
	opLeU:        "leu",
	opAddF:       "addf",
	opSubF:       "subf",
	opMulF:       "mulf",
	opDivF:       "divf",
	opNegF:       "negf",
	opEqF:        "eqf",
	opNeF:        "nef",
	opLtF:        "ltf",
	opLeF:        "lef",
	opI2F:        "i2f",
	opU2F:        "u2f",
	opNotB:       "notb",
	opIsNil:      "isnil",
	opEqR:        "eqr",
	opNeR:        "ner",
	opEqV:        "eqv",
	opConcat:     "concat",
	opEqS:        "eqs",
	opNeS:        "nes",
	opLtS:        "lts",
	opLeS:        "les",
	opJmp:        "jmp",
	opLoop:       "loop",
	opJmpT:       "jmpt",
	opJmpF:       "jmpf",
	opJNltI:      "jnlti",
	opJNleI:      "jnlei",
	opJNeqI:      "jneqi",
	opJNneI:      "jnnei",
	opJNltF:      "jnltf",
	opJNleF:      "jnlef",
	opCall:       "call",
	opCallNative: "callnative",
	opCallHost:   "callhost",
	opCallValue:  "callvalue",
	opRet:        "ret",
	opLLen:       "llen",
	opLGetN:      "lgetn",
	opLGetS:      "lgets",
	opLGetR:      "lgetr",
	opLGet:       "lget",
	opLGetOk:     "lgetok",
	opLAppN:      "lappn",
	opLAppS:      "lapps",
	opLAppR:      "lappr",
	opLApp:       "lapp",
	opLNew:       "lnew",
	opLSet:       "lset",
	opBox:        "box",
	opZeroR:      "zeror",
}

// Dump returns the bytecode of the program, e.g. for a test.
func (p *Program) Dump() string {
	var sb strings.Builder
	for i, t := range p.types {
		fmt.Fprintf(&sb, "type %d %s kind=%d l=%v obj=%v\n", i, t.name, t.kind, t.l, t.obj)
	}
	for _, f := range append(p.funcs, p.init) {
		fmt.Fprintf(&sb, "func %s frame=%v params=%v results=%v\n", f.name, f.frame, f.params.l, f.results.l)
		for pc, in := range f.code {
			name := fmt.Sprintf("op%d", in.op)
			if int(in.op) < len(opNames) {
				name = opNames[in.op]
			}
			fmt.Fprintf(&sb, "  %3d %-10s %d %d %d", pc, name, in.a, in.b, in.c)
			switch in.op {
			case opCall, opCallNative, opCallHost, opCallValue:
				cs := f.calls[in.a]
				name := ""
				switch {
				case cs.fn != nil:
					name = cs.fn.name
				case cs.nat != nil:
					name = cs.nat.name
				case cs.ext != nil:
					name = cs.ext.name
				}
				fmt.Fprintf(&sb, "\t; %s base=%v", name, cs.base)
			case opLdkS:
				fmt.Fprintf(&sb, "\t; %q", p.sconst[in.b])
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
