// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package task names work, which ends after the call, which started it, e.g. a network request of lib/net. The layers
// of a program exchange only its [ID]: a use case starts the work and returns the ID as a plain value, and a view, which
// may not be imported by the use case, waits for it with ui.OnDone.
//
//gostrict:header
package task

// ID names a started job. It is a plain value without any meaning, valid only in the window, which started the job.
type ID string
