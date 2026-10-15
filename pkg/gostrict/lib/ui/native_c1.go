// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	"time"

	"go.wdy.de/nago/presentation/ui/alert"
)

// alertOptionsOf returns the options of nago of the options, without the zero options, which have no value.
func alertOptionsOf(opts []AlertOption) []alert.Option {
	res := make([]alert.Option, 0, len(opts))
	for _, o := range opts {
		if opt := get[alert.Option](o.ref); opt != nil {
			res = append(res, opt)
		}
	}
	return res
}

func (m AlertMessage) nago() alert.Message {
	return alert.Message{Title: m.Title, Message: m.Message, Intent: alert.Intent(m.Intent), Duration: time.Duration(m.Duration) * time.Millisecond}
}

func alertMessageFrom(m alert.Message) AlertMessage {
	return AlertMessage{Title: m.Title, Message: m.Message, Intent: AlertIntent(m.Intent), Duration: int32(m.Duration / time.Millisecond)}
}
