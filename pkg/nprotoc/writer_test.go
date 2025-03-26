// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

import (
	"testing"
)

func Test_parseFieldHeader(t *testing.T) {
	// apply an exhaustive test for all shift combinations
	for si := range 8 {
		for fi := range 32 {
			fh := fieldHeader{
				shape:   shape(si),
				fieldId: fieldId(fi),
			}

			if v := parseFieldHeader(fh.asValue()); v != fh {
				t.Errorf("parseFieldHeader mismatch: got %v, want %v", v, fh)
			}
		}
	}
}
