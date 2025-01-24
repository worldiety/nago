package nprotoc

func (c *Compiler) tsEmitEnum(t Typename, decl Enum) {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("interface %s {\n", t)
	c.inc()
	c.pf(c.makeGoDoc("a marker method to indicate the enum / union type membership"))
	c.pf("is%s(): void;\n", t)
	c.dec()
	c.pn("}")
}
