package nprotoc

func (c *Compiler) goEmitArray(t Typename, decl Array) error {
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("type %s []%s\n", t, decl.Type)
	c.pn("")
	c.pf("func(v *%s) write(w *BinaryWriter)error{\n", t)
	c.pn("	if err:=w.writeUvarint(uint64(len(*v)));err!=nil {")
	c.pn("		return err")
	c.pn("	}")
	c.pn("	for _,item:=range *v {")
	c.pn("		if err:=item.writeTypeHeader(w);err!=nil {")
	c.pn("			return err")
	c.pn("		}")
	c.pn("		if err:=item.write(w);err!=nil {")
	c.pn("			return err")
	c.pn("		}")
	c.pn("  }")
	c.pn("	return nil")
	c.pn("}\n")

	c.pf("func(v *%s) read(r *BinaryReader)error{\n", t)
	c.pn("	count,err := r.readUvarint()")
	c.pn("	if err != nil {")
	c.pn("		return err")
	c.pn("	}")
	c.pn("")
	c.pf("	slice:=make([]%s, count)\n", decl.Type)
	c.pn("	for i:=uint64(0); i<count; i++ {")
	c.pn("		obj, err:=Unmarshal(r)")
	c.pn("		if err != nil {")
	c.pn("			return err")
	c.pn("		}")
	c.pf("      slice[i]=obj.(%s)\n", decl.Type)
	c.pn("  }")
	c.pn("")
	c.pf("	*v = slice\n")
	c.pn("	return nil")
	c.pn("}\n")

	c.pf("func(v *%s) IsZero()bool{\nreturn v==nil || *v==nil || len(*v)==0\n}\n\n", t)

	c.pf("func(v *%s) reset(){\n*v=nil\n}\n\n", t)

	return nil
}
