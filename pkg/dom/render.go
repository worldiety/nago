// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import (
	"html"
	"io"
	"sort"
)

// renderElement writes a complete HTML element (open tag, children, close tag)
// or a void element (self-closing open tag) to w.
func renderElement(tag string, attrs map[string]string, kids []Node, void bool, w io.Writer) error {
	if _, err := io.WriteString(w, "<"+tag); err != nil {
		return err
	}
	// write attributes in sorted order for deterministic output
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := attrs[k]
		escaped := html.EscapeString(v)
		if _, err := io.WriteString(w, ` `+k+`="`+escaped+`"`); err != nil {
			return err
		}
	}
	if void {
		_, err := io.WriteString(w, ">")
		return err
	}
	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}
	for _, child := range kids {
		if err := child.Render(w); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "</"+tag+">")
	return err
}
