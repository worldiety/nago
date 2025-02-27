package nprotoc

func (c *Compiler) tsEmitBool(t Typename, decl Bool) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export type %s = boolean \n", t)

	if err := c.tsEmitWriteTypeHeader(t); err != nil {
		return err
	}
	
	return nil
}
