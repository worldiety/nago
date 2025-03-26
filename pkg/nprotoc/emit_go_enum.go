// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

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
	buf.WriteString("\tIsZero() bool\n")
	buf.WriteString("\treset()\n")
	buf.WriteString("\tWriteable\n")
	buf.WriteString("\tReadable\n")
	buf.WriteString("}\n")

	buf.WriteString("\n")

	for _, variant := range decl.Variants {
		buf.WriteString(fmt.Sprintf("func(%s)is%s(){}\n", variant, t))
	}

	c.marshals = append(c.marshals, buf.String())
}
