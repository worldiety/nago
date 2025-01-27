package nprotoc

import (
	"bytes"
	"fmt"
)

func (c *Compiler) emitMarshal() error {
	var buf bytes.Buffer

	buf.WriteString("type Writeable interface {\nwrite(*BinaryWriter)error\nwriteTypeHeader(*BinaryWriter)error}\n")
	buf.WriteString("func Marshal(dst *BinaryWriter, src Writeable)error{\n")
	buf.WriteString("if err:=src.writeTypeHeader(dst);err!=nil{\nreturn err\n}\n")
	buf.WriteString("if err:=src.write(dst);err!=nil{\nreturn err\n}\n")
	buf.WriteString("return nil\n}\n")
	c.marshals = append(c.marshals, buf.String())

	for typename, decl := range c.sortedDecl() {
		idDecl, ok := decl.(IdentityTypeDeclaration)
		if !ok {
			continue
		}
		
		if err := c.goEmitWriteTypeHeader(typename, idDecl); err != nil {
			return err
		}

	}

	return nil
}

func (c *Compiler) emitUnmarshal() error {
	var buf bytes.Buffer

	buf.WriteString("type Readable interface {\nread(*BinaryReader)error\n}\n")
	buf.WriteString("func Unmarshal(src *BinaryReader)(Readable,error){\n")
	buf.WriteString("_,tid,err := src.readTypeHeader()\nif err!=nil{\nreturn nil,err\n}\n")
	buf.WriteString("switch tid {\n")
	for typename, decl := range c.sortedDecl() {
		idDecl, ok := decl.(IdentityTypeDeclaration)
		if !ok {
			continue
		}

		buf.WriteString(fmt.Sprintf("case %d:\n", idDecl.ID()))
		buf.WriteString(fmt.Sprintf("var v %s\n", typename))
		buf.WriteString("if err:=v.read(src);err!=nil{\nreturn nil,err\n}\n")
		buf.WriteString("return &v,nil\n")

	}

	buf.WriteString("default:\nreturn nil, fmt.Errorf(\"unknown type in marshal: %d\", tid)\n")
	buf.WriteString("}\n")
	buf.WriteString("\n}\n")
	c.marshals = append(c.marshals, buf.String())

	return nil
}

func (c *Compiler) goEmitWriteTypeHeader(t Typename, decl IdentityTypeDeclaration) error {
	c.pf("func(v *%s) writeTypeHeader(w *BinaryWriter)error{\n", t)
	c.inc()
	sh, err := c.shapeOf(t)
	if err != nil {
		return err
	}
	c.pf("if err:=w.writeTypeHeader(%s, %d);err!=nil{\nreturn err\n}\n", sh, decl.ID())
	c.dec()
	c.pn("return nil")
	c.pn("}\n")

	return nil
}
