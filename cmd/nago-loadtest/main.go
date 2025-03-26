// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"go.wdy.de/nago/testing"
)

func main() {
	err := testing.NewTester().Test()
	if err != nil {
		fmt.Println(err)
	}
}
