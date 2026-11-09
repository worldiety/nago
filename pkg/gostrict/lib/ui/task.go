// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/task"
	"go.wdy.de/nago/pkg/gostrict/std/task/taskhost"
)

// OnDone calls fn on the event loop of the window, like a click, when the job has ended and its work was committed:
// failure is "" on success and otherwise the reason, e.g. for a state, which renders the window again. Call it in the
// same call, which started the job. A job holds one function, so another OnDone replaces it. Closing the window drops
// the function, but not the job. An unknown job calls fn with a failure.
//
//gostrict:host
//gostrict:calls later
func OnDone(wnd Window, id task.ID, fn func(failure string)) {
	taskhost.Of(wnd.nago()).Wait(id, fn)
}
