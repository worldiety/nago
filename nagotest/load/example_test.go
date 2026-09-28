// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package load_test

import (
	"testing"

	"go.wdy.de/nago/nagotest/load"
)

// the example of the command must stay valid
func TestExampleFile(t *testing.T) {
	if _, err := load.LoadFile("../../cmd/nago-loadtest/example.json"); err != nil {
		t.Fatal(err)
	}
}
