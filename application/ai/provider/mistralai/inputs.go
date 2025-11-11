// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import "go.wdy.de/nago/pkg/xjson"

var inputVariants = []xjson.VariantOption{
	xjson.Variant[MessageInputEntry]("message.input"),
}

type InputBox struct {
	Values []Entry
}

func (c InputBox) MarshalJSON() ([]byte, error) {
	return xjson.MarshalInternally("type", c.Values, inputVariants...)
}

func (c *InputBox) UnmarshalJSON(data []byte) error {
	tmp, err := xjson.UnmarshalInternally("type", data, inputVariants...)
	if err != nil {
		return err
	}
	c.Values = nil
	switch t := tmp.(type) {
	case []any:
		for _, a := range t {
			c.Values = append(c.Values, a.(Entry))
		}
	default:
		c.Values = append(c.Values, tmp.(Entry))
	}

	return nil
}
