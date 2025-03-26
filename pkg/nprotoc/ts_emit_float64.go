// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

func (c *Compiler) tsEmitFloat64(t Typename, decl Float64) error {
	c.pn("")
	c.p(c.makeGoDoc(decl.Doc))
	c.pf("export type %s = number\n", t)

	if err := c.tsEmitWriteTypeHeader(t); err != nil {
		return err
	}
	
	return nil
}
