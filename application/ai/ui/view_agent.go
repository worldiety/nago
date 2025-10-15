// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
)

type ViewAgent struct {
	ID           agent.ID           `visible:"false"`
	Name         string             ``
	Description  string             ``
	Prompt       string             ``
	Model        agent.Model        ``
	Libraries    []library.ID       ``
	Capabilities []agent.Capability ``
	UserEditable bool               ``
}

type TAgent struct {
	ag *core.State[agent.Agent]
}

func Agent(agent *core.State[agent.Agent]) TAgent {
	return TAgent{
		ag: agent,
	}
}

func (c TAgent) Render(ctx core.RenderContext) core.RenderNode {
	return form.Auto[agent.Agent](form.AutoOptions{Window: ctx.Window()}, c.ag).Render(ctx)
}
