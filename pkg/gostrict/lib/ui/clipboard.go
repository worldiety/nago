// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

// ClipboardSetText asks the browser of the window to write the text into the clipboard of the user, while the window
// renders as well as in a function, which the window calls later. An empty text does nothing, so it never clears the
// clipboard, and there is no limit of the length. Like core.Window.Clipboard of nago, the program does not learn,
// whether the browser has written the text: if the browser refuses, the window shows a dialog with a button, which
// writes it with the click of the user. There is no function to read the clipboard.
//
//gostrict:host
func ClipboardSetText(wnd Window, text string) {
	if text == "" {
		return
	}

	// SetText only publishes the event to the browser and always returns nil
	_ = wnd.nago().Clipboard().SetText(text)
}
