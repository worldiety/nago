// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

func (c *Compiler) tsEmitUint(t Typename, decl Uint) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export type %s = number \n", t)

	if err := c.tsEmitWriteTypeHeader(t); err != nil {
		return err
	}

	if len(decl.ConstValues) > 0 {
		c.pf("// companion enum containing all defined constants for %s\n", t)
		c.pf("export enum %sValues {\n", t)
		c.inc()
		for value, con := range decl.sortedConst() {
			c.i()
			c.p(c.makeGoDoc(con.Doc))
			c.pf("%s = %s,\n", con.Name, value)
		}
		c.dec()
		c.pn("}\n\n")
	}

	return nil
}
