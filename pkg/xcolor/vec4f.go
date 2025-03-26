// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xcolor

type Vec4f [4]float32

type Vec4i8 [4]uint8

func fTU8(v float32) uint8 {
	x := max(0, min(1, v))
	return uint8(x * 255)
}
