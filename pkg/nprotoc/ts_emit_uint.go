package nprotoc

func (c *Compiler) tsEmitUint(t Typename, decl Uint) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export class %s %s{\n", t, c.tsImplements(t))
	c.inc()
	c.p(` 
	private value: number; // Using number to handle uint64 (precision limits apply)
	
	constructor(value: number = 0) {
		this.value = value;
	}
	
	isZero(): boolean {
		return this.value === 0;
	}
	
	reset(): void {
		this.value = 0;
	}
	
	write(writer: BinaryWriter): void {
		writer.writeUvarint(this.value);
	}
	
	read(reader: BinaryReader): void {
		this.value = reader.readUvarint();
	}

`)

	if err := c.tsEmitWriteTypeHeader(t); err != nil {
		return err
	}

	c.dec()

	c.pf("}\n\n")

	if len(decl.ConstValues) > 0 {
		c.pf("// companion enum containing all defined constants for %s\n", t)
		c.pf("enum %sValues {\n", t)
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
