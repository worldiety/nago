// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package adm

import (
	"fmt"
	"github.com/worldiety/enum/json"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

var counter atomic.Int64

func Write(dir string, cmd Command) error {
	name := fmt.Sprintf("%s-%d.json", time.Now().Format("2006-01-02-15-04-05.000"), counter.Add(1))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_ = os.MkdirAll(dir, 0755)
	}

	buf, err := json.MarshalFor[Command](cmd)
	if err != nil {
		return err
	}

	fname := filepath.Join(dir, name)
	return os.WriteFile(fname, buf, 0644)
}
