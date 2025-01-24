package nprotoc

import "strings"

func (c *Compiler) tsEmitRecord(t Typename, decl Record) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("class %s %s{\n", t, c.tsImplements(t))
	c.inc()

	for _, field := range decl.sortedFields() {
		c.pf(c.makeGoDoc(field.Doc))
		c.pf("%s: %s;\n\n", tsFieldName(field.Name), field.Type)
	}

	c.tsEmitRecordConstructor(t, decl)
	if err := c.tsEmitRecordRead(t, decl); err != nil {
		return err
	}

	if err := c.tsEmitRecordWrite(t, decl); err != nil {
		return err
	}

	for _, s := range c.tsMarkerMethods(t) {
		c.pn(s)
	}
	c.dec()
	c.pn("}\n")

	return nil
}

func (c *Compiler) tsEmitRecordConstructor(t Typename, decl Record) {
	c.i()
	c.p("constructor(")
	for _, field := range decl.sortedFields() {
		c.p(tsFieldName(field.Name), ": ", string(field.Type))
		c.p(" = new ", string(field.Type), "(), ")
	}
	c.p(") {\n")
	c.inc()
	for _, field := range decl.sortedFields() {
		c.pf("this.%s = %[1]s;\n", tsFieldName(field.Name))
	}
	c.dec()
	c.pn("}\n")
}

func (c *Compiler) tsEmitRecordRead(t Typename, decl Record) error {
	c.pn("read(reader: BinaryReader): void {")
	c.inc()

	for _, field := range decl.sortedFields() {
		c.pf("this.%s.reset();\n", tsFieldName(field.Name))
	}

	c.pn("const fieldCount = reader.readByte();")
	c.pn("for (let i = 0; i < fieldCount; i++) {")
	c.inc()
	c.pn("const fieldHeader = reader.readFieldHeader();")
	c.pn("switch (fieldHeader.fieldId) {")
	c.inc()
	for fid, field := range decl.sortedFields() {
		c.pf("case %d:\n", fid)
		c.inc()
		c.pf("this.%s.read(reader);\n", tsFieldName(field.Name))
		c.pn("break")
		c.dec()
	}

	c.pn("default:")
	c.inc()
	c.pn("throw new Error(`Unknown field ID: ${fieldHeader.fieldId}`);")
	c.dec()

	c.dec()
	c.pn("}")
	c.dec()
	c.pn("}")
	c.dec()
	c.pn("}\n")

	return nil
}

func (c *Compiler) tsEmitRecordWrite(t Typename, decl Record) error {

	c.pn("write(writer: BinaryWriter): void {")
	c.inc()

	c.i()
	c.p("const fields = [false,")
	for _, field := range decl.sortedFields() {
		c.p("!this.", tsFieldName(field.Name), ".isZero(),")
	}
	c.p("];\n")

	c.pn("let fieldCount = fields.reduce((count, present) => count + (present ? 1 : 0), 0);")
	c.pn("writer.writeByte(fieldCount);")

	for fid, field := range decl.sortedFields() {
		c.pf("if (fields[%d]) {\n", fid)
		c.inc()
		sh, err := c.shapeOf(field.Type)
		if err != nil {
			return err
		}
		c.pf("writer.writeFieldHeader(Shapes.%s, %d);\n", strings.ToUpper(sh.String()), fid)
		c.pf("this.%s.write(writer);\n", tsFieldName(field.Name))
		c.dec()
		c.pn("}")
	}

	c.dec()
	c.pn("}\n")

	return nil
}
