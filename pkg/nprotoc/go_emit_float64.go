// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

func (c *Compiler) goEmitFloat64(t Typename, decl Float64) error {
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("type %s float64\n", t)
	c.pn("")
	c.pf("func(v *%s) write(r *BinaryWriter)error{\n", t)
	c.pn("	return r.writeFloat64(float64(*v))")
	c.pn("}\n")

	c.pf("func(v *%s) read(r *BinaryReader)error{\n", t)
	c.pn("	val,err := r.readFloat64()")
	c.pn("	if err != nil {")
	c.pn("		return err")
	c.pn("	}")
	c.pn("")
	c.pf("  *v=%s(val)\n", t)
	c.pn("	return nil")
	c.pn("}\n")

	c.pf("func(v *%s) IsZero()bool{\nreturn *v==0.0\n}\n\n", t)

	c.pf("func(v *%s) reset(){\n*v=0.0\n}\n\n", t)

	return nil
}
