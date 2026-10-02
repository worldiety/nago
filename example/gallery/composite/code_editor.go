// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
)

const codeSample = `package main

import "fmt"

func main() {
	for i := range 3 {
		fmt.Println("hello", i)
	}
}
`

func init() {
	app.Register("code-editor", func(wnd core.Window) core.View {
		source := core.AutoState[string](wnd).Init(func() string { return codeSample })

		return CodeEditor(source.Get()).
			InputValue(source).
			Language("go").
			Frame(Frame{Width: L560, Height: L256})
	})
}
