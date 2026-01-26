// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"context"
	"fmt"

	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/pkg/xslices"
)

type FormTextStyle string

const (
	FormTextStyleDefault  FormTextStyle = "text"
	FormTextStyleH1       FormTextStyle = "h1"
	FormTextStyleH2       FormTextStyle = "h2"
	FormTextStyleH3       FormTextStyle = "h3"
	FormTextStyleMarkdown FormTextStyle = "markdown"
)

var FormTextValues = xslices.Wrap(FormTextStyleDefault, FormTextStyleH1, FormTextStyleH2, FormTextStyleH3, FormTextStyleMarkdown)

type FormTextAdded struct {
	Workspace WorkspaceID   `json:"workspace"`
	Form      FormID        `json:"form"`
	Parent    ViewID        `json:"parent"`
	After     ViewID        `json:"after,omitempty"` // optional, if empty add as first element (even if not empty), otherwise after (usually below or right of)
	ID        ViewID        `json:"id"`
	Value     string        `json:"value"`
	Style     FormTextStyle `json:"style"`
}

func (evt FormTextAdded) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormTextAdded) Discriminator() evs.Discriminator {
	return "FormTextAdded"
}

func (evt FormTextAdded) Evolve(ctx context.Context, ws *Workspace) error {
	form, ok := GetViewGroup(ws, evt.Form, evt.Parent)
	if !ok {
		return fmt.Errorf("form %s not found", evt.Form)
	}

	form.Insert(NewFormText(evt.ID, evt.Value, evt.Style), evt.After)
	return nil
}
