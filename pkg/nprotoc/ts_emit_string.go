package nprotoc

func (c *Compiler) tsEmitString(t Typename, decl String) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export type %s = string\n", t)

	if err := c.tsEmitWriteTypeHeader(t); err != nil {
		return err
	}

	return nil
}
