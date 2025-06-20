// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
)

func (c *Compiler) emitGoInt(t Typename, decl Int) {
	var buf bytes.Buffer
	buf.WriteString(c.makeGoDoc(decl.Doc))
	buf.WriteString(fmt.Sprintf("type %s int64\n", t))
	if len(decl.ConstValues) > 0 {
		buf.WriteString("const (\n")
		keys := slices.Sorted(maps.Keys(decl.ConstValues)) // provide stable order
		for _, lit := range keys {
			con := decl.ConstValues[lit]
			buf.WriteString(c.makeGoDoc(con.Doc))
			buf.WriteString(fmt.Sprintf("\t%s %s = %s\n", con.Name, t, lit))
		}
		buf.WriteString(")\n")
	}

	buf.WriteString(fmt.Sprintf("func(v *%s) write(r *BinaryWriter)error{\n", t))
	buf.WriteString("return r.writeVarint(int64(*v))\n")
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func(v *%s) read(r *BinaryReader)error{\n", t))
	buf.WriteString(fmt.Sprintf("tmp,err:=r.readVarint()\n"))
	buf.WriteString(fmt.Sprintf("if err!=nil{\n"))
	buf.WriteString(fmt.Sprintf("return err\n"))
	buf.WriteString("}\n")
	buf.WriteString(fmt.Sprintf("*v=%s(tmp)\n", t))
	buf.WriteString(fmt.Sprintf("return nil\n"))
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func(v *%s) reset(){\n*v=%s(0)\n}\n", t, t))

	buf.WriteString(fmt.Sprintf("func(v *%s) IsZero()bool{\nreturn*v==0\n}\n", t))

	c.marshals = append(c.marshals, buf.String())
}
