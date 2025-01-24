package nprotoc

import (
	"bytes"
	"fmt"
)

func (c *Compiler) emitGoEnum(t Typename, decl Enum) {
	var buf bytes.Buffer
	buf.WriteString(c.makeGoDoc(decl.Doc))
	buf.WriteString(fmt.Sprintf("type %s interface {\n", t))
	buf.WriteString(c.makeGoDoc("a marker method to indicate the enum / union type membership"))
	buf.WriteString(fmt.Sprintf("\tis%s()\n", t))
	buf.WriteString("}\n")

	buf.WriteString("\n")

	for _, variant := range decl.Variants {
		buf.WriteString(fmt.Sprintf("func(%s)is%s(){}\n", variant, t))
	}

	c.marshals = append(c.marshals, buf.String())
}
