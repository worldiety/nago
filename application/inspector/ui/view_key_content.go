// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiinspector

import (
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/application/inspector"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"strings"
)

func viewKeyContent(wnd core.Window, uc inspector.UseCases, store *core.State[inspector.Store], entryState *core.State[inspector.PageEntry]) core.View {
	if entryState.Get().Key == "" {
		return nil
	}

	var body core.View

	var textState *core.State[string]

	entry := entryState.Get()
	if entry.MimeType == "application/json" {
		notJson := false
		var text string
		optBuf, err := blob.Get(store.Get().Store, entry.Key)
		if err != nil {
			text = err.Error()
		} else {
			if optBuf.IsSome() {
				var tmp map[string]any
				if err := json.Unmarshal(optBuf.Unwrap(), &tmp); err != nil {
					// probably not json or damaged
					// just continue
					notJson = true
				} else {
					buf, err := json.MarshalIndent(tmp, "", "  ")
					if err != nil {
						text = err.Error()
					} else {
						text = string(buf)
					}
				}

			}
		}

		if !notJson {
			textState = getTextState(wnd, entryState).Init(func() string {
				return text
			})

			body = ui.CodeEditor(textState.Get()).InputValue(textState).Language("json")
		}

	} else if strings.HasPrefix(entry.MimeType, "text/") {
		var text string
		optBuf, err := blob.Get(store.Get().Store, entry.Key)
		if err != nil {
			text = err.Error()
		} else {
			if optBuf.IsSome() {
				text = string(optBuf.Unwrap())
			}
		}

		textState = getTextState(wnd, entryState).Init(func() string {
			return text
		})

		body = ui.CodeEditor(textState.Get()).InputValue(textState).Frame(ui.Frame{}.FullWidth())
	}

	if body != nil {

		// this is the text edit
		return ui.ScrollView(body).Axis(ui.ScrollViewAxisBoth).Frame(ui.Frame{Width: ui.Full, Height: cssHeight})
	}

	buf, err := blob.Get(store.Get().Store, entry.Key)

	return ui.VStack(
		ui.IfFunc(err != nil, func() core.View {
			return ui.Text(err.Error())
		}),
		ui.Text(entry.MimeType),
		ui.Text(fmt.Sprintf("%d bytes", len(buf.UnwrapOr(nil)))),
		ui.TertiaryButton(func() {
			wnd.ExportFiles(core.ExportFilesOptions{
				Files: []core.File{
					core.MemFile{
						Filename:     entry.Key,
						MimeTypeHint: entry.MimeType,
						Bytes:        buf.UnwrapOr(nil),
					},
				},
			})
		}).PreIcon(flowbiteOutline.Download),
	).Gap(ui.L8).FullWidth()

}

func getTextState(wnd core.Window, entryState *core.State[inspector.PageEntry]) *core.State[string] {
	return core.StateOf[string](wnd, "editor-text-"+entryState.Get().Key)
}
