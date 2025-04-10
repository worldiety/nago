// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

func (c *Compiler) goEmitMap(t Typename, decl Map) error {
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("type %s map[%s]%s\n", t, decl.Key, decl.Value)
	c.pn("")
	c.pf("func(v *%s) write(w *BinaryWriter)error{\n", t)
	c.pn("	if err:=w.writeUvarint(uint64(len(*v)));err!=nil {")
	c.pn("		return err")
	c.pn("	}")
	c.pn("	for k,v:=range *v {")
	c.pn("  	// key")
	c.pn("		if err:=k.writeTypeHeader(w);err!=nil {")
	c.pn("			return err")
	c.pn("		}")
	c.pn("		if err:=k.write(w);err!=nil {")
	c.pn("			return err")
	c.pn("		}")
	c.pn("  	// value")
	c.pn("		if err:=v.writeTypeHeader(w);err!=nil {")
	c.pn("			return err")
	c.pn("		}")
	c.pn("		if err:=v.write(w);err!=nil {")
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
	c.pf("	tmpMap:=make(map[%s]%s, count) \n", decl.Key, decl.Value)
	c.pn("	for i:=uint64(0); i<count; i++ {")
	c.pn("		k, err:=Unmarshal(r)")
	c.pn("		if err != nil {")
	c.pn("			return err")
	c.pn("		}")
	c.pn("		v, err:=Unmarshal(r)")
	c.pn("		if err != nil {")
	c.pn("			return err")
	c.pn("		}")
	c.pf("      tmpMap[*k.(*%s)]=*v.(*%s)\n", decl.Key, decl.Value)
	c.pn("  }")
	c.pn("")
	c.pf("	*v = tmpMap\n")
	c.pn("	return nil")
	c.pn("}\n")

	c.pf("func(v *%s) IsZero()bool{\nreturn v==nil || *v==nil || len(*v)==0\n}\n\n", t)

	c.pf("func(v *%s) reset(){\n*v=nil\n}\n\n", t)

	return nil
}
