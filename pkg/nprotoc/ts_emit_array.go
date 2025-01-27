package nprotoc

func (c *Compiler) tsEmitArray(t Typename, decl Array) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("class %s %s {\n", t, c.tsImplements(t))
	c.inc()
	c.pf("private value: %s[];\n", decl.Type)
	c.pn("")
	c.pf("constructor(value: %s[] = []) {\n", decl.Type)
	c.p(` 
      this.value = value;
    }

  isZero(): boolean {
	return !this.value || this.value.length === 0;
  }

  reset(): void {
	this.value = [];
  }


  write(writer: BinaryWriter): void {
	writer.writeUvarint(this.value.length); // Write the length of the array
	for (const c of this.value) {
	  c.writeTypeHeader(writer); // Write the type header for each component
	  c.write(writer); // Write the component data
	}
  }

  read(reader: BinaryReader): void {
	const count = reader.readUvarint(); // Read the length of the array
	const components: Component[] = [];

	for (let i = 0; i < count; i++) {
	  const obj = unmarshal(reader); // Read and unmarshal each component
	  components.push(obj as Component); // Cast and add to the array
	}

	this.value = components;
  }

`)

	if err := c.tsEmitWriteTypeHeader(t); err != nil {
		return err
	}

	c.dec()

	c.pf("}\n\n")

	return nil
}
