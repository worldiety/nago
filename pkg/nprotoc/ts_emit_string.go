package nprotoc

func (c *Compiler) tsEmitString(t Typename, decl String) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export class %s %s{\n", t, c.tsImplements(t))
	c.inc()
	c.p(` 
  private value: string; 

  constructor(value: string = "") {
    this.value = value;
  }

  isZero(): boolean {
    return this.value === "";
  }

  reset(): void {
    this.value = "";
  }

  // Get the string representation of the Color
  toString(): string {
    return this.value;
  }

  write(writer: BinaryWriter): void {
    const data = new TextEncoder().encode(this.value); // Convert string to Uint8Array
    writer.writeUvarint(data.length); // Write the length of the string
    writer.write(data); // Write the string data
  }

  read(reader: BinaryReader): void {
	const strLen = reader.readUvarint(); // Read the length of the string
    const buf = reader.readBytes(strLen); // Read the string data
    this.value = new TextDecoder().decode(buf); // Convert Uint8Array to string
  }

`)

	if err := c.tsEmitWriteTypeHeader(t); err != nil {
		return err
	}

	c.dec()
	c.pf("}\n\n")

	return nil
}
