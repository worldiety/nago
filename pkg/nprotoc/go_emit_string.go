package nprotoc

func (c *Compiler) goEmitString(t Typename, decl String) error {
	strShape := "string"
	if decl.Go.Type != "" {
		strShape = decl.Go.Type
	}

	c.p(c.makeGoDoc(decl.Doc))
	c.pf("type %s %s\n", t, strShape)
	c.pn("")
	c.pf("func(v *%s) write(r *BinaryWriter)error{\n", t)
	c.pn("	data := *(*[]byte)(unsafe.Pointer(v))")
	c.pn("	if err:=r.writeUvarint(uint64(len(data)));err!=nil {")
	c.pn("		return err")
	c.pn("	}")
	c.pn("	return r.write(data)")
	c.pn("}\n")

	c.pf("func(v *%s) read(r *BinaryReader)error{\n", t)
	c.pn("	strLen,err := r.readUvarint()")
	c.pn("	if err != nil {")
	c.pn("		return err")
	c.pn("	}")
	c.pn("")
	c.pn("	buf:=make([]byte, strLen)\n")
	c.pn("	if err:=r.read(buf);err!=nil{")
	c.pn("		return err")
	c.pn("	}")
	c.pn("")
	c.pf("	*v = *(*%s)(unsafe.Pointer(&buf))\n", t)
	c.pn("	return nil")
	c.pn("}\n")

	c.pf("func(v *%s) IsZero()bool{\nreturn len(*v)==0\n}\n\n", t)

	c.pf("func(v *%s) reset(){\n*v=%[1]s(\"\")\n}\n\n", t)

	return nil
}
