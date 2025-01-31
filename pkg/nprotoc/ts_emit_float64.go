package nprotoc

func (c *Compiler) tsEmitFloat64(t Typename, decl Float64) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export class %s %s{\n", t, c.tsImplements(t))
	c.inc()
	c.p(` 
	public value: number; 
	
	constructor(value: number = 0.0) {
		this.value = value;
	}
	
	isZero(): boolean {
		return this.value === 0.0;
	}
	
	reset(): void {
		this.value = 0.0;
	}
	
	write(writer: BinaryWriter): void {
		writer.writeFloat64(this.value);
	}
	
	read(reader: BinaryReader): void {
		this.value = reader.readFloat64();
	}

`)

	if err := c.tsEmitWriteTypeHeader(t); err != nil {
		return err
	}

	c.dec()

	c.pf("}\n\n")

	return nil
}
