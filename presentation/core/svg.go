// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

// SVG contains the valid embeddable source of Scalable Vector Graphics.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type SVG []byte

func (svg SVG) AsBytes() []byte {
	return svg
}

func (svg SVG) Empty() bool {
	return len(svg) == 0
}
