// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagogen

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"testing/fstest"

	"go.wdy.de/nago/app/builder/aam"
)

func genMod(fsys fstest.MapFS, model *aam.App) error {
	vname := "v0.0.0"
	if info, ok := debug.ReadBuildInfo(); ok {
		vname = info.Main.Version
		vname = strings.TrimSuffix(vname, "+dirty")
	}
	modName := strings.TrimPrefix(string(model.GitRepoURL), "https://")
	modName = strings.TrimPrefix(modName, "http://")
	modName = strings.TrimPrefix(modName, "git@")
	modName = strings.ReplaceAll(modName, ":", "/")
	buf := []byte(fmt.Sprintf(`module %s

go 1.25

require (
	go.wdy.de/nago %s
)`, modName, vname))

	fsys["go.mod"] = &fstest.MapFile{Data: buf, Mode: os.ModePerm}
	return nil
}
