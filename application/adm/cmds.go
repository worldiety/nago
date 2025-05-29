// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package adm

import (
	"github.com/worldiety/enum"
	"time"
)

var CommandEnum = enum.Declare[Command,
	func(
		func(EnableBootstrapAdmin),
		func(any),
	),
]()

type Command interface {
	isCommand()
}

type EnableBootstrapAdmin struct {
	AliveUntil time.Time `json:"aliveUntil"`
	Password   string    `json:"password"`
}

func (EnableBootstrapAdmin) isCommand() {}
