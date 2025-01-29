package nprotoc

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
)

func (c *Compiler) emitGoRecord(t Typename, decl Record) error {
	// TODO optimize for zero field types
	var buf bytes.Buffer
	buf.WriteString(c.makeGoDoc(decl.Doc))
	buf.WriteString(fmt.Sprintf("type %s struct {\n", t))
	keys := slices.Sorted(maps.Keys(decl.Fields)) // provide stable order
	for _, fid := range keys {
		field := decl.Fields[fid]
		buf.WriteString(c.makeGoDoc(field.Doc))
		buf.WriteString(fmt.Sprintf("\t%s %s\n", field.Name, field.Type))
	}
	buf.WriteString("}\n")

	// write
	buf.WriteString(fmt.Sprintf("func(v *%s) write(w *BinaryWriter)error{\n", t))
	buf.WriteString(fmt.Sprintf("var fields [%d]bool\n", len(decl.Fields)+1))
	for _, fid := range keys {
		field := decl.Fields[fid]
		sh, err := c.shapeOf(field.Type)
		if err != nil {
			return err
		}

		if sh == xobjectAsArray {
			buf.WriteString(fmt.Sprintf("fields[%d] = v.%[2]s != nil && !v.%[2]s.IsZero()\n", fid, field.Name))
		} else {
			buf.WriteString(fmt.Sprintf("fields[%d] = !v.%s.IsZero()\n", fid, field.Name))
		}
	}

	buf.WriteString(`
fieldCount:=byte(0)
for _,present:= range fields {
	if present {
		fieldCount++
	}
}
`)

	buf.WriteString("if err:=w.writeByte(fieldCount); err != nil {\nreturn err\n}\n")

	for _, fid := range keys {
		field := decl.Fields[fid]
		buf.WriteString(fmt.Sprintf("if fields[%d]{\n", fid))
		sh, err := c.shapeOf(field.Type)
		if err != nil {
			return fmt.Errorf("shape of field %s is unknown: %w", field.Type, err)
		}
		isVirtualObject := sh == xobjectAsArray
		if isVirtualObject {
			sh = array
			buf.WriteString("// polymorphic field (enum) type encodes as polymorphic array\n")
		}
		buf.WriteString(fmt.Sprintf("if err:=w.writeFieldHeader(%s,%d);err!=nil{\nreturn err\n}\n", sh.String(), fid))
		if isVirtualObject {
			buf.WriteString("if err:=w.writeUvarint(1);err!=nil{\nreturn err\n}\n")
			buf.WriteString(fmt.Sprintf("if err:=v.%s.writeTypeHeader(w);err!=nil{\nreturn err\n}\n", field.Name))
		}

		buf.WriteString(fmt.Sprintf("if err:=v.%s.write(w);err!=nil{\nreturn err\n}\n", field.Name))

		buf.WriteString("}\n")
	}

	buf.WriteString("return nil\n")
	buf.WriteString("}\n\n")
	// read

	buf.WriteString(fmt.Sprintf("func(v *%s) read(r *BinaryReader)error{\n", t))

	buf.WriteString(fmt.Sprintf("v.reset()\n"))

	buf.WriteString("fieldCount,err:=r.readByte()\n")
	buf.WriteString("if err!=nil{\nreturn err\n}\n")
	buf.WriteString("for range fieldCount {\n")
	buf.WriteString("fh,err:=r.readFieldHeader()\n")
	buf.WriteString("if err!=nil{\nreturn err\n}\n")
	buf.WriteString("switch fh.fieldId {\n")
	for _, fid := range keys {
		field := decl.Fields[fid]
		sh, err := c.shapeOf(field.Type)
		if err != nil {
			return err
		}

		switch sh {
		case xobjectAsArray:
			buf.WriteString(fmt.Sprintf("\tcase %d:\n", fid))
			buf.WriteString("// polymorphic field type (enum) decodes as polymorphic array\n")
			buf.WriteString("count,err:=r.readUvarint()\n if err!=nil{\nreturn err\n}\n")
			buf.WriteString("if count!=1{\nreturn fmt.Errorf(\"expected exact 1 element in enum field\")\n}\n")

			buf.WriteString("obj,err:=Unmarshal(r)\n if err!=nil{\nreturn err\n}\n")
			buf.WriteString(fmt.Sprintf("v.%s=obj.(%s)\n", field.Name, field.Type))
		default:
			buf.WriteString(fmt.Sprintf("\tcase %d:\n", fid))
			buf.WriteString(fmt.Sprintf("err:=v.%s.read(r)\n", field.Name))
			buf.WriteString("if err!=nil{\nreturn err\n}\n")
		}

	}
	buf.WriteString("}\n")
	buf.WriteString("}\n")
	buf.WriteString(fmt.Sprintf("return nil\n"))
	buf.WriteString("}\n")

	buf.WriteString("\n")

	c.marshals = append(c.marshals, buf.String())

	if err := c.goEmitRecordReset(t, decl); err != nil {
		return err
	}

	if err := c.goEmitRecordIsZero(t, decl); err != nil {
		return err
	}

	return nil
}

func (c *Compiler) goEmitRecordReset(t Typename, decl Record) error {
	c.pf("func(v *%s) reset() {\n", t)
	c.inc()

	for _, field := range decl.sortedFields() {
		sh, err := c.shapeOf(field.Type)
		if err != nil {
			return err
		}

		if sh == xobjectAsArray {
			c.pf("v.%s=nil\n", goFieldName(field.Name))
		} else {
			c.pf("v.%s.reset()\n", goFieldName(field.Name))
		}

	}

	c.dec()
	c.pn("}\n")

	return nil
}

func (c *Compiler) goEmitRecordIsZero(t Typename, decl Record) error {
	c.pf("func(v *%s) IsZero() bool {\n", t)
	c.inc()

	c.i()
	c.p("return ")
	idx := 0
	for _, field := range decl.sortedFields() {
		c.p("v.", goFieldName(field.Name), ".IsZero()")
		if idx < decl.fieldCount()-1 {
			c.p(" && ")
		}
		idx++
	}

	if len(decl.Fields) == 0 {
		c.p(" true")
	}
	c.p("\n")

	c.dec()
	c.pn("}\n")

	return nil
}
