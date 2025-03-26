// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

func (c *Compiler) goEmitBool(t Typename, decl Bool) error {
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("type %s bool\n", t)
	c.pn("")
	c.pf("func(v *%s) write(r *BinaryWriter)error{\n", t)
	c.pn("	val:=uint64(0)")
	c.pn("	if v!=nil && *v {")
	c.pn("		val=1")
	c.pn("	}")
	c.pn("	return r.writeUvarint(val)")
	c.pn("}\n")

	c.pf("func(v *%s) read(r *BinaryReader)error{\n", t)
	c.pn("	val,err := r.readUvarint()")
	c.pn("	if err != nil {")
	c.pn("		return err")
	c.pn("	}")
	c.pn("")
	c.pn("	if val==1 {")
	c.pn("		*v=true")
	c.pn("	} else {")
	c.pn("		*v=false")
	c.pn("	}")
	c.pn("")
	c.pn("	return nil")
	c.pn("}\n")

	c.pf("func(v *%s) IsZero()bool{\nreturn *v==false\n}\n\n", t)

	c.pf("func(v *%s) reset(){\n*v=false\n}\n\n", t)

	return nil
}
