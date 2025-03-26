// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

func (c *Compiler) tsEmitArray(t Typename, decl Array) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export class %s %s {\n", t, c.tsImplements(t))
	c.inc()
	c.pf("public value: %s[];\n", decl.Type)
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
`)
	if c.isPrimitive(decl.Type) {
		c.pf("		writeTypeHeader%s(writer)\n", decl.Type)
	} else {
		c.pf("c.writeTypeHeader(writer); // Write the type header for each component)\n")
	}

	if c.isString(decl.Type) {
		c.pf("		writeString(writer, c)\n")
	} else if c.isBool(decl.Type) {
		c.pf("		writeBool(writer, c)\n")
	} else if c.isInt(decl.Type) {
		c.pf("		writeInt(writer, c)\n")
	} else if c.isFloat(decl.Type) {
		c.pf("		writeFloat(writer, c)\n")
	} else {
		c.pf("		c.write(writer); // Write the component data\n")
	}

	c.p(`
	  //c.writeTypeHeader(writer); // Write the type header for each component
	  //c.write(writer); // Write the component data
	}
  }


`)

	c.pf(`  
  read(reader: BinaryReader): void {
	const count = reader.readUvarint(); // Read the length of the array
	const values: %[1]s[] = [];

	for (let i = 0; i < count; i++) {
	  const obj = unmarshal(reader); // Read and unmarshal each component
	  values.push((obj as any) as %[1]s); // Cast and add to the array
	}

	this.value = values;
  }`, decl.Type)
	c.pn("")

	if err := c.tsEmitWriteTypeHeaderMethod(t); err != nil {
		return err
	}

	c.dec()

	c.pf("}\n\n")

	return nil
}
