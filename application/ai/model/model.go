// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package model

type ID string
type Model struct {
	ID                 ID      `json:"id,omitempty"`
	Name               string  `json:"name,omitempty"`
	Description        string  `json:"description,omitempty"`
	DefaultTemperature float64 `json:"default_temperature,omitempty"`
}

func (m Model) Identity() ID {
	return m.ID
}

func (m Model) WithIdentity(id ID) Model {
	m.ID = id
	return m
}

func (m Model) String() string {
	return m.Name + "\n" + m.Description
}
