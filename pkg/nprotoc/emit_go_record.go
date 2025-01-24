package nprotoc

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
)

func (c *Compiler) emitGoRecord(t Typename, decl Record) error {
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
		buf.WriteString(fmt.Sprintf("fields[%d] = !v.%s.IsZero()\n", fid, field.Name))
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
		buf.WriteString(fmt.Sprintf("if err:=w.writeFieldHeader(%s,%d);err!=nil{\nreturn err\n}\n", sh.String(), fid))
		buf.WriteString(fmt.Sprintf("if err:=v.%s.write(w);err!=nil{\nreturn err\n}\n", field.Name))
		buf.WriteString("}\n")
	}

	buf.WriteString("return nil\n")
	buf.WriteString("}\n\n")
	// read

	buf.WriteString(fmt.Sprintf("func(v *%s) read(r *BinaryReader)error{\n", t))

	for _, fid := range keys {
		field := decl.Fields[fid]
		buf.WriteString(fmt.Sprintf("v.%s.reset()\n", field.Name))
	}

	buf.WriteString("fieldCount,err:=r.readByte()\n")
	buf.WriteString("if err!=nil{\nreturn err\n}\n")
	buf.WriteString("for range fieldCount {\n")
	buf.WriteString("fh,err:=r.readFieldHeader()\n")
	buf.WriteString("if err!=nil{\nreturn err\n}\n")
	buf.WriteString("switch fh.fieldId {\n")
	for _, fid := range keys {
		field := decl.Fields[fid]
		buf.WriteString(fmt.Sprintf("\tcase %d:\n", fid))
		buf.WriteString(fmt.Sprintf("err:=v.%s.read(r)\n", field.Name))
		buf.WriteString("if err!=nil{\nreturn err\n}\n")
	}
	buf.WriteString("}\n")
	buf.WriteString("}\n")
	buf.WriteString(fmt.Sprintf("return nil\n"))
	buf.WriteString("}\n")

	buf.WriteString("\n")

	buf.WriteString(fmt.Sprintf("func(v *%s) IsZero()bool{\nreturn *v==(%s{})\n}\n", t, t))

	c.marshals = append(c.marshals, buf.String())

	return nil
}
