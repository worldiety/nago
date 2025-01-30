package nprotoc

func (c *Compiler) tsEmitMap(t Typename, decl Map) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export class %s %s {\n", t, c.tsImplements(t))
	c.inc()
	c.pf("private value: Map<%s,%s>;\n", decl.Key, decl.Value)
	c.pn("")
	c.pf("constructor(value: Map<%[1]s,%[2]s> = new Map<%[1]s,%[2]s>()) {\n", decl.Key, decl.Value)
	c.pf(` 
      this.value = value;
    }

  isZero(): boolean {
	return !this.value || this.value.size === 0;
  }

  reset(): void {
	this.value = new Map<%[1]s,%[2]s>();
  }


  write(writer: BinaryWriter): void {
	writer.writeUvarint(this.value.size); // Write the length of the map
	for (const [key, value] of this.value) {
      // write key
	  key.writeTypeHeader(writer); 
	  key.write(writer); 

      // write value
	  value.writeTypeHeader(writer); 
	  value.write(writer); 
	}
  }

`, decl.Key, decl.Value)

	c.pf(`  
  read(reader: BinaryReader): void {
	const count = reader.readUvarint(); 
	const values = new Map<%[1]s,%[2]s>();

	for (let i = 0; i < count; i++) {
	  const key = unmarshal(reader);
      const val = unmarshal(reader);

	  values.set(key as %[1]s, val as %[2]s); 
	}

	this.value = values;
  }`, decl.Key, decl.Value)
	c.pn("")

	if err := c.tsEmitWriteTypeHeader(t); err != nil {
		return err
	}

	c.dec()

	c.pf("}\n\n")

	return nil
}
