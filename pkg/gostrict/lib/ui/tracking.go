// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/ui/tracking"
)

// TTrackingSupportRequestDialog is an overlay, which allows the user to contact the developers about an unexpected
// error, see TrackingSupportRequestDialog.
type TTrackingSupportRequestDialog struct{ ref js.Ref }

func (TTrackingSupportRequestDialog) isView() {}

// TrackingSupportRequestDialog returns the dialog, which shows the latest anonymous error code, or nothing, if there
// is no unhandled error, like tracking.SupportRequestDialog of nago.
//
//gostrict:host
func TrackingSupportRequestDialog(wnd Window) TTrackingSupportRequestDialog {
	return TTrackingSupportRequestDialog{ref: wrap(tracking.SupportRequestDialog(wnd.nago()))}
}
