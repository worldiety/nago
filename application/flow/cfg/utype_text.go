// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgflow

import (
	"encoding/json"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"golang.org/x/text/language"
)

var (
	StrTextLabel       = i18n.MustString("nago.flow.utype.text.label", i18n.Values{language.English: "Text", language.German: "Text"})
	StrTextDescription = i18n.MustString("nago.flow.utype.text.desc", i18n.Values{language.English: "Text represents the underlying type for single line text fields.", language.German: "Text stellt den unterliegenden Typ für einzeilige Textfelder dar."})
)

// Text is a single-line text field type.
type Text struct {
}

func (s Text) ID() TID {
	return "text"
}

func (s Text) Default() string {
	return ""
}

func (s Text) Label() i18n.StrHnd {
	return StrTextLabel
}

func (s Text) Description() i18n.StrHnd {
	return StrTextDescription
}

func (s Text) Render(info FieldInfo, state *core.State[string]) core.View {
	return ui.TextField(info.Label, state.Get()).
		ID(info.ID).
		SupportingText(info.SupportingText).
		InputValue(state)
}

func (s Text) Validate(t string) error {
	return nil
}

func (s Text) Marshal(t string) (json.RawMessage, error) {
	return json.Marshal(t)
}

func (s Text) Unmarshal(message json.RawMessage) (string, error) {
	var v string
	if err := json.Unmarshal(message, &v); err != nil {
		return "", err
	}

	return v, nil
}
