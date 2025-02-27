package nprotoc

import "fmt"

func (c *Compiler) tsEmitMap(t Typename, decl Map) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export class %s %s {\n", t, c.tsImplements(t))
	c.inc()
	c.pf("public value: Map<%s,%s>;\n", decl.Key, decl.Value)
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
	  //key.writeTypeHeader(writer); 
	  //key.write(writer); 

 


`, decl.Key, decl.Value)

	if !c.isPrimitive(decl.Key) {
		return fmt.Errorf("code generator does only support primitive keys")
	}
	c.pf("  writeTypeHeader%s(writer);\n", decl.Key)
	c.pf("  writeString(writer,key);\n")

	c.pn("         // write value")
	if c.isPrimitive(decl.Value) {
		c.pf("  writeTypeHeader%s(writer);\n", decl.Value)

		if c.isString(decl.Value) {
			c.pf("  writeString(writer,value);\n")
		} else if c.isInt(decl.Value) {
			c.pf("  writeInt(writer,value);\n")
		} else if c.isFloat(decl.Value) {
			c.pf("  writeFloat(writer,value);\n")
		} else if c.isBool(decl.Value) {
			c.pf("  writeBool(writer,value);\n")
		} else {
			return fmt.Errorf("code generator does only support primitive values: %v", decl.Value)
		}

	} else {
		c.p(`
	  value.writeTypeHeader(writer); 
	  value.write(writer); 
`)
	}

	c.p(`

	}
  }
`)

	c.pf(`  
  read(reader: BinaryReader): void {
	const count = reader.readUvarint(); 
	const values = new Map<%[1]s,%[2]s>();

	for (let i = 0; i < count; i++) {
	  const key = unmarshal(reader);
      const val = unmarshal(reader);

	  values.set((key as any) as %[1]s, (val as any) as %[2]s); 
	}

	this.value = values;
  }`, decl.Key, decl.Value)
	c.pn("")

	if err := c.tsEmitWriteTypeHeaderMethod(t); err != nil {
		return err
	}

	c.dec()

	c.pf("}\n\n")

	return nil
}
