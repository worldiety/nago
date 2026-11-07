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

// Main reports "value <error>" for an error of target, which it handles as a value, else "result <results>". Beside
// the error it may not read the results of target (wokoda R-SCRIPT-GS-OPTIONAL); n10/pass passes them on to the host.
func Main(id string, reject bool) (string, error) {
	info, err := target.Main(id, reject)
	if err != nil {
		return "value " + err.Error(), nil
	}
	return "result " + info.Name + "/" + text.FromInt(info.N) + "/" + text.FromInt(info.Tags.Len()), nil
}
