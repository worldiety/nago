// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

func (c *Compiler) tsEmitMarshal() error {
	c.pn("// Function to marshal a Writeable object into a BinaryWriter")
	c.pn("export function marshal(dst: BinaryWriter, src: Writeable): void {")
	c.inc()
	c.pn("src.writeTypeHeader(dst);")
	c.pn("src.write(dst);")
	c.dec()
	c.pn("}\n")

	return nil
}

func (c *Compiler) tsEmitUnmarshal() error {
	c.pn("// Function to unmarshal data from a BinaryReader into a Readable object")
	c.pn("export function unmarshal(src: BinaryReader): any {")
	c.inc()
	c.pn("const { typeId } = src.readTypeHeader();")
	c.pn("switch (typeId) {")
	c.inc()
	for typename, decl := range c.sortedDecl() {
		id, ok := decl.(IdentityTypeDeclaration)
		if !ok {
			continue
		}

		c.pf("case %d: {\n", id.ID())
		c.inc()
		if c.isString(typename) {
			c.pf("const v = readString(src) as %s;\n", typename)
		} else if c.isBool(typename) {
			c.pf("const v = readBool(src) as %s;\n", typename)
		} else if c.isSint(typename) {
			c.pf("const v = readSint(src) as %s;\n", typename)
		} else if c.isUint(typename) {
			c.pf("const v = readInt(src) as %s;\n", typename)
		} else if c.isFloat(typename) {
			c.pf("const v = readFloat(src) as %s;\n", typename)
		} else {
			c.pf("const v = new %s();\n", typename)
			c.pn("v.read(src);")
		}

		c.pn("return v;")
		c.dec()
		c.pn("}")
	}
	c.dec()
	c.pn("}")
	c.pn("throw new Error(`Unknown type ID: ${typeId}`);")
	c.dec()
	c.pn("}\n")

	return nil
}
