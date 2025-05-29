// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package adm

import (
	"github.com/worldiety/enum/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type ReadCommandsOptions struct {
	DeleteAfterRead bool
}

func ReadCommands(dir string, opts ReadCommandsOptions) []Command {
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		slog.Error("failed to read adm cmd directory", "dir", dir, "err", err.Error())
		return nil
	}

	var cmds []Command
	for _, file := range files {
		// we are sorted by filename, thus ordering is already correct
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		buf, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			slog.Error("failed to read adm cmd file", "file", file.Name(), "err", err.Error(), "dir", dir)
			continue
		}

		var cmd Command
		if err := json.Unmarshal(buf, &cmd); err != nil {
			slog.Error("failed to parse adm cmd json file", "file", file.Name(), "err", err.Error(), "dir", dir)
			continue
		}

		if opts.DeleteAfterRead {
			if err := os.Remove(filepath.Join(dir, file.Name())); err != nil {
				slog.Error("failed to delete adm cmd file", "file", file.Name(), "err", err.Error())
			}
		}

		cmds = append(cmds, cmd)
	}

	return cmds
}
