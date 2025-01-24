package nprotoc

import "strings"

func (c *Compiler) tsEmitMarshal() error {
	c.pn("// Function to marshal a Writeable object into a BinaryWriter")
	c.pn("function marshal(dst: BinaryWriter, src: Writeable): void {")
	c.inc()
	for typename, decl := range c.sortedDecl() {
		id, ok := decl.(IdentityTypeDeclaration)
		if !ok {
			continue
		}

		c.pf("if (src instanceof %s) {\n", typename)
		c.inc()
		sh, err := c.shapeOf(typename)
		if err != nil {
			return err
		}
		c.pf("dst.writeTypeHeader(Shapes.%s, %d);\n", strings.ToUpper(sh.String()), id.ID())
		c.pn("src.write(dst);")
		c.pn("return")
		c.dec()
		c.pn("}")
	}
	c.dec()
	c.pn("}\n")

	return nil
}

func (c *Compiler) tsEmitUnmarshal() error {
	c.pn("// Function to unmarshal data from a BinaryReader into a Readable object")
	c.pn("function unmarshal(src: BinaryReader): Readable {")
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
		c.pf("const v = new %s();\n", typename)
		c.pn("v.read(src);")
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
