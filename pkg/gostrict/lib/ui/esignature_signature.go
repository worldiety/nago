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
	"go.wdy.de/nago/presentation/ui/esignature"
)

// TEsignatureSignature is a signature field with a top and a bottom text, like esignature.TSignature of nago. See
// EsignatureSignature.
type TEsignatureSignature struct{ ref js.Ref }

func (TEsignatureSignature) isView() {}

// EsignatureSignature creates an empty signature, like esignature.Signature of nago.
//
//gostrict:host
func EsignatureSignature() TEsignatureSignature {
	return TEsignatureSignature{ref: wrap(esignature.Signature())}
}

// Frame sets the frame of the signature container.
//
//gostrict:host
func (c TEsignatureSignature) Frame(frame Frame) TEsignatureSignature {
	return TEsignatureSignature{ref: wrap(get[esignature.TSignature](c.ref).Frame(frame.nago()))}
}

// TopText sets the text displayed above the signature body.
//
//gostrict:host
func (c TEsignatureSignature) TopText(text string) TEsignatureSignature {
	return TEsignatureSignature{ref: wrap(get[esignature.TSignature](c.ref).TopText(text))}
}

// BottomText sets the text displayed below the signature body.
//
//gostrict:host
func (c TEsignatureSignature) BottomText(text string) TEsignatureSignature {
	return TEsignatureSignature{ref: wrap(get[esignature.TSignature](c.ref).BottomText(text))}
}

// Body sets the main body view of the signature.
//
//gostrict:host
func (c TEsignatureSignature) Body(v View) TEsignatureSignature {
	return TEsignatureSignature{ref: wrap(get[esignature.TSignature](c.ref).Body(viewOf(v)))}
}
