// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package permission_test

import (
	"go.wdy.de/nago/application/permission"
	"testing"
)

func TestRegister(t *testing.T) {
	permission.Register[MakeStuff](permission.Permission{ID: "de.worldiety.test"})
	//permission.Make[MakeStuff]("de.worldiety.test")
}

type MakeStuff func()
