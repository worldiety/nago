// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

func (c *Compiler) tsEmitEnum(t Typename, decl Enum) {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export interface %s extends Writeable, Readable{\n", t)
	c.inc()
	c.p(c.makeGoDoc("a marker method to indicate the enum / union type membership"))
	c.pf("is%s(): void;\n", t)
	c.dec()
	c.pn("}")
}
