package nprotoc

func (c *Compiler) tsEmitBool(t Typename, decl Bool) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("class %s %s{\n", t, c.tsImplements(t))
	c.inc()
	c.p(` 
	private value: boolean; 
	
	constructor(value: boolean = false) {
		this.value = value;
	}
	
	isZero(): boolean {
		return !this.value;
	}
	
	reset(): void {
		this.value = false;
	}
	
	write(writer: BinaryWriter): void {
		writer.writeUvarint(this.value?1:0);
	}
	
	read(reader: BinaryReader): void {
		this.value = reader.readUvarint() === 1;
	}

`)

	if err := c.tsEmitWriteTypeHeader(t); err != nil {
		return err
	}

	c.dec()

	c.pf("}\n\n")

	return nil
}
