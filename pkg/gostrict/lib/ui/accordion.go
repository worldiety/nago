// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/ui/accordion"
)

// TAccordion is an expandable section, which shows or hides its body, see Accordion.
type TAccordion struct{ ref js.Ref }

func (TAccordion) isView() {}

// Accordion returns an accordion with the header and the body, which is shown, while open is true. Like
// accordion.Accordion of nago.
//
//gostrict:host
func Accordion(header, body View, open BoolState) TAccordion {
	return TAccordion{ref: wrap(accordion.Accordion(viewOf(header), viewOf(body), open.nago()))}
}

// Frame sets the frame to control the bounds of the accordion.
//
//gostrict:host
func (c TAccordion) Frame(frame Frame) TAccordion {
	return TAccordion{ref: wrap(get[accordion.TAccordion](c.ref).Frame(frame.nago()))}
}

// Small enables a smaller visual appearance of the accordion.
//
//gostrict:host
func (c TAccordion) Small() TAccordion {
	return TAccordion{ref: wrap(get[accordion.TAccordion](c.ref).Small())}
}

// HideSeparator hides the bottom border (separator) of the accordion.
//
//gostrict:host
func (c TAccordion) HideSeparator() TAccordion {
	return TAccordion{ref: wrap(get[accordion.TAccordion](c.ref).HideSeparator())}
}

// InputValue binds the accordion to an external state, which controls whether it is open.
//
//gostrict:host
func (c TAccordion) InputValue(input BoolState) TAccordion {
	return TAccordion{ref: wrap(get[accordion.TAccordion](c.ref).InputValue(input.nago()))}
}

// FullWidth sets the width of the accordion to the full width of the parent.
//
//gostrict:host
func (c TAccordion) FullWidth() TAccordion {
	return TAccordion{ref: wrap(get[accordion.TAccordion](c.ref).FullWidth())}
}
