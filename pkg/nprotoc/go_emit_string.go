package nprotoc

func (c *Compiler) goEmitString(t Typename, decl String) error {
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("type %s string\n", t)
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
	c.pn("}")
	return nil
}
