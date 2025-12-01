// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagogen

import (
	"bytes"
	"os"
	"testing/fstest"

	gen "github.com/dave/jennifer/jen"
	"go.wdy.de/nago/app/builder/aam"
)

func genMain(fsys fstest.MapFS, model *aam.App) error {
	f := gen.NewFile("main")
	f.Func().Id("main").Params().Block(
		gen.Qual("fmt", "Println").Call(gen.Lit("Hello, world")),
	)

	var buf bytes.Buffer
	if err := f.Render(&buf); err != nil {
		return err
	}

	fsys["cmd/"+model.ID.Last()+"-srv/main.go"] = &fstest.MapFile{
		Mode: os.ModePerm,
		Data: buf.Bytes(),
	}

	return nil
}
