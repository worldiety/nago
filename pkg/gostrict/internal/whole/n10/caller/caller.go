// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package caller calls target and reports its results, also an error, which it handles as a value.
package caller

import (
	"go.wdy.de/nago/pkg/gostrict/internal/whole/n10/target"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Main reports "value <error> zero <results>" for an error of target, else "result <results>".
func Main(id string, reject bool) (string, error) {
	info, err := target.Main(id, reject)
	got := info.Name + "/" + text.FromInt(info.N) + "/" + text.FromInt(info.Tags.Len())
	if err != nil {
		return "value " + err.Error() + " zero " + got, nil
	}
	return "result " + got, nil
}
