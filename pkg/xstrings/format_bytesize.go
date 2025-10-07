// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xstrings

import (
	"fmt"

	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

func FormatByteSize(tag language.Tag, bytes int64, decimals int) string {
	const (
		_           = iota
		KiB float64 = 1 << (10 * iota)
		MiB
		GiB
		TiB
	)

	b := float64(bytes)

	switch {
	case b >= TiB:
		return i18n.FormatFloat(tag, float64(bytes)/TiB, decimals, "TiB")
	case b >= GiB:
		return i18n.FormatFloat(tag, float64(bytes)/GiB, decimals, "GiB")
	case b >= MiB:
		return i18n.FormatFloat(tag, float64(bytes)/MiB, decimals, "MiB")
	case b >= KiB:
		return i18n.FormatFloat(tag, float64(bytes)/KiB, decimals, "KiB")
	default:
		return fmt.Sprintf("%d B", bytes)
	}

}
