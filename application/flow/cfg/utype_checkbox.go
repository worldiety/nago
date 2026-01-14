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
	StrCheckboxLabel       = i18n.MustString("nago.flow.utype.checkbox.label", i18n.Values{language.English: "Checkbox", language.German: "Ankreuzfeld"})
	StrCheckboxDescription = i18n.MustString("nago.flow.utype.checkbox.desc", i18n.Values{language.English: "Checkbox represents the underlying type for single line text fields.", language.German: "Ankreuzfelder stellen den unterliegenden Typ für Ankreuzoptionen dar."})
)

// Checkbox is a boolish checkbox.
type Checkbox struct {
}

func (s Checkbox) ID() TID {
	return "checkbox"
}

func (s Checkbox) Default() bool {
	return false
}

func (s Checkbox) Label() i18n.StrHnd {
	return StrCheckboxLabel
}

func (s Checkbox) Description() i18n.StrHnd {
	return StrCheckboxDescription
}

func (s Checkbox) Render(info FieldInfo, state *core.State[bool]) core.View {
	return ui.CheckboxField(info.Label, state.Get()).
		ID(info.ID).
		InputValue(state).
		SupportingText(info.SupportingText)
}

func (s Checkbox) Validate(t bool) error {
	return nil
}

func (s Checkbox) Marshal(t bool) (json.RawMessage, error) {
	return json.Marshal(t)
}

func (s Checkbox) Unmarshal(message json.RawMessage) (bool, error) {
	var v bool
	if err := json.Unmarshal(message, &v); err != nil {
		return false, err
	}

	return v, nil
}
