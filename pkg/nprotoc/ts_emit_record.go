package nprotoc

import "strings"

func (c *Compiler) tsEmitRecord(t Typename, decl Record) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export class %s %s {\n", t, c.tsImplements(t))
	c.inc()

	for _, field := range decl.sortedFields() {
		c.pf(c.makeGoDoc(field.Doc))
		if c.tsCanBeUndefined(field.Type) {
			c.pf("public %s?: %s;\n\n", tsFieldName(field.Name), field.Type)
		} else {
			c.pf("public %s: %s;\n\n", tsFieldName(field.Name), field.Type)
		}

	}

	c.tsEmitRecordConstructor(t, decl)
	if err := c.tsEmitRecordRead(t, decl); err != nil {
		return err
	}

	if err := c.tsEmitRecordWrite(t, decl); err != nil {
		return err
	}

	if err := c.tsEmitRecordIsZero(t, decl); err != nil {
		return err
	}

	if err := c.tsEmitRecordReset(t, decl); err != nil {
		return err
	}

	if err := c.tsEmitWriteTypeHeaderMethod(t); err != nil {
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
		if c.tsCanBeUndefined(field.Type) {
			c.p(tsFieldName(field.Name), ": ")
			c.p(string(field.Type), "|undefined")
			c.p(" = undefined, ")
		} else {
			c.p(tsFieldName(field.Name), ": ", string(field.Type))
			c.p(" = new ", string(field.Type), "(), ")
		}

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
	if len(decl.Fields) == 0 {
		c.pn("read(r: BinaryReader): void {r.readByte();}")
		return nil
	}

	c.pn("read(reader: BinaryReader): void {")
	c.inc()

	c.pn("this.reset();")

	c.pn("const fieldCount = reader.readByte();")
	c.pn("for (let i = 0; i < fieldCount; i++) {")
	c.inc()
	c.pn("const fieldHeader = reader.readFieldHeader();")
	c.pn("switch (fieldHeader.fieldId) {")
	c.inc()
	for fid, field := range decl.sortedFields() {
		c.pf("case %d: {\n", fid)
		c.inc()
		sh, err := c.shapeOf(field.Type)
		if err != nil {
			return err
		}
		if sh == xobjectAsArray {
			c.pn("// decode polymorphic field as 1 element array")
			c.pn("const len = reader.readUvarint();")
			c.pn("if (len != 1) {")
			c.pn("  throw new Error(`unexpected length: ` + len);")
			c.pn("}")
			c.pf("this.%s = unmarshal(reader) as %s;\n", tsFieldName(field.Name), field.Type)
		} else {
			if c.isString(field.Type) {
				c.pf("this.%s = readString(reader);\n", tsFieldName(field.Name))
			} else if c.isFloat(field.Type) {
				c.pf("this.%s = readFloat(reader);\n", tsFieldName(field.Name))
			} else if c.isBool(field.Type) {
				c.pf("this.%s = readBool(reader);\n", tsFieldName(field.Name))
			} else if c.isInt(field.Type) {
				c.pf("this.%s = readInt(reader);\n", tsFieldName(field.Name))
			} else {
				c.pf("this.%s=new %s()\n", tsFieldName(field.Name), field.Type)
				c.pf("this.%s.read(reader);\n", tsFieldName(field.Name))
			}
		}

		c.pn("break")
		c.dec()
		c.pn("}")
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
	if len(decl.Fields) == 0 {
		c.pn("write(w: BinaryWriter): void {w.writeByte(0);}")
		return nil
	}

	c.pn("write(writer: BinaryWriter): void {")
	c.inc()

	c.i()
	c.p("const fields = [false,")
	for _, field := range decl.sortedFields() {
		if c.tsCanBeUndefined(field.Type) {
			c.p("this.", tsFieldName(field.Name), "!== undefined")
			if !c.isPrimitive(field.Type) {
				c.p(" && !this.", tsFieldName(field.Name), ".isZero(),")
			} else {
				c.p(",")
			}
		} else {
			c.p("!this.", tsFieldName(field.Name), ".isZero(),")
		}
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

		if sh == xobjectAsArray {
			c.pn("// encode polymorphic enum as 1 element slice")
			c.pf("writer.writeFieldHeader(Shapes.%s, %d);\n", strings.ToUpper(array.String()), fid)
			c.pf("writer.writeByte(1);\n")
		} else {
			c.pf("writer.writeFieldHeader(Shapes.%s, %d);\n", strings.ToUpper(sh.String()), fid)
		}

		if c.tsCanBeUndefined(field.Type) {
			if c.isString(field.Type) {
				c.pf("writeString(writer,this.%s!); // typescript linters cannot see, that we already checked this properly above\n", tsFieldName(field.Name))
			} else if c.isFloat(field.Type) {
				c.pf("writeFloat(writer,this.%s!); // typescript linters cannot see, that we already checked this properly above\n", tsFieldName(field.Name))
			} else if c.isBool(field.Type) {
				c.pf("writeBool(writer,this.%s!); // typescript linters cannot see, that we already checked this properly above\n", tsFieldName(field.Name))
			} else if c.isInt(field.Type) {
				c.pf("writeInt(writer,this.%s!); // typescript linters cannot see, that we already checked this properly above\n", tsFieldName(field.Name))
			} else {
				c.pf("this.%s!.write(writer); // typescript linters cannot see, that we already checked this properly above\n", tsFieldName(field.Name))
			}
		} else {
			c.pf("this.%s.write(writer);\n", tsFieldName(field.Name))
		}
		c.dec()
		c.pn("}")
	}

	c.dec()
	c.pn("}\n")

	return nil
}

func (c *Compiler) tsEmitRecordIsZero(t Typename, decl Record) error {
	if len(decl.Fields) == 0 {
		c.pn("isZero(): boolean {return true;}")
		return nil
	}

	c.pn("isZero(): boolean {")
	c.inc()

	c.i()
	c.p("return ")
	idx := 0
	for _, field := range decl.sortedFields() {
		if c.tsCanBeUndefined(field.Type) {
			c.p("(this.", tsFieldName(field.Name), " === undefined ")
			if c.isPrimitive(field.Type) {
				c.p(")")
			} else {
				c.p("|| this.", tsFieldName(field.Name), ".isZero())")
			}
		} else {
			c.p("this.", tsFieldName(field.Name), ".isZero()")
		}
		if idx < decl.fieldCount()-1 {
			c.p(" && ")
		}
		idx++
	}
	c.p("\n")

	c.dec()
	c.pn("}\n")

	return nil
}

func (c *Compiler) tsEmitRecordReset(t Typename, decl Record) error {
	if len(decl.Fields) == 0 {
		c.pn("reset(): void {}")
		return nil
	}

	c.pn("reset(): void {")
	c.inc()

	for _, field := range decl.sortedFields() {
		if c.tsCanBeUndefined(field.Type) {
			c.pf("this.%s = undefined\n", tsFieldName(field.Name))
		} else {
			c.pf("this.%s.reset()\n", tsFieldName(field.Name))
		}
	}

	c.dec()
	c.pn("}\n")

	return nil
}
