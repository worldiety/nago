// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgflow

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"golang.org/x/text/language"
)

var (
	StrTextAreaLabel       = i18n.MustString("nago.flow.utype.textarea.label", i18n.Values{language.English: "Text", language.German: "Text"})
	StrTextAreaDescription = i18n.MustString("nago.flow.utype.textarea.desc", i18n.Values{language.English: "Text represents the underlying type for single line text fields.", language.German: "Text stellt den unterliegenden Typ für einzeilige Textfelder dar."})
)

// TextArea is a multi-line text field type.
type TextArea struct {
	Text
}

func (s TextArea) ID() TID {
	return "textarea"
}

func (s TextArea) Label() i18n.StrHnd {
	return StrTextAreaLabel
}

func (s TextArea) Description() i18n.StrHnd {
	return StrTextAreaDescription
}

func (s TextArea) Render(info FieldInfo, state *core.State[string]) core.View {
	return ui.TextField(info.Label, state.Get()).
		ID(info.ID).
		SupportingText(info.SupportingText).
		InputValue(state).Lines(5)
}
