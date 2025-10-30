// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package agent

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xtime"
)

type ID string

// Temperature defines how reproducible the answers of the Agent are. Lower values make the results more predictable.
// Higher values make the results more novel. The allowed range is [0..1].
type Temperature float64
type Agent struct {
	ID           ID                     `json:"id,omitempty"`
	Name         string                 `json:"name,omitempty"`
	Description  string                 `json:"desc,omitempty"`
	Instructions string                 `json:"prompt,omitempty"`
	Model        model.ID               `json:"model,omitempty"`
	Libraries    []library.ID           `json:"libraries,omitempty"`
	Temperature  Temperature            `json:"tmp,omitempty"`
	UpdatedAt    xtime.UnixMilliseconds `json:"lastMod,omitempty"`
	CreatedBy    user.ID                `json:"createdBy,omitempty"`
	CreatedAt    xtime.UnixMilliseconds `json:"createdAt,omitempty"`
}

func (e Agent) Identity() ID {
	return e.ID
}

type CreateOptions struct {
	Name         string
	Description  string
	Model        model.ID
	Instructions string
}

type UpdateOptions struct {
	Name         option.Opt[string]
	Description  option.Opt[string]
	Model        option.Opt[model.ID]
	Instructions option.Opt[string]
	Temperature  option.Opt[Temperature]
	Libraries    option.Opt[[]library.ID]
}

type Repository data.Repository[Agent, ID]
