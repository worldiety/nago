// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"fmt"
	"slices"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Agent = (*mistralAgent)(nil)

type mistralAgent struct {
	parent *mistralProvider
	id     agent.ID
}

func (p *mistralAgent) Identity() agent.ID {
	return p.id
}

func (p *mistralAgent) Update(subject auth.Subject, opts agent.UpdateOptions) (agent.Agent, error) {
	ag, err := p.client().GetAgent(string(p.id))
	if err != nil {
		return agent.Agent{}, fmt.Errorf("cannot load existing agent: %w", err)
	}

	if opts.Name.IsSome() {
		ag.Name = opts.Name.Unwrap()
	}

	if opts.Description.IsSome() {
		ag.Description = opts.Description.Unwrap()
	}

	if opts.Temperature.IsSome() {
		tmp := float64(opts.Temperature.Unwrap())
		ag.CompletionArgs.Temperature = &tmp
	}

	if opts.Model.IsSome() {
		ag.Model = string(opts.Model.Unwrap())
	}

	if opts.Instructions.IsSome() {
		ag.Instructions = opts.Instructions.Unwrap()
	}

	if opts.Libraries.IsSome() {
		ag.Tools = slices.DeleteFunc(ag.Tools, func(tool Tool) bool {
			return tool.Type == ToolDocumentLibrary
		})

		if len(opts.Libraries.Unwrap()) > 0 {
			ag.Tools = append(ag.Tools, Tool{
				Type:      ToolDocumentLibrary,
				Libraries: opts.Libraries.Unwrap(),
			})
		}
	}

	if opts.Tools.IsSome() {
		ag.Tools = slices.DeleteFunc(ag.Tools, func(tool Tool) bool {
			return tool.Type != ToolDocumentLibrary // the rest is just build-in
		})

		for _, id := range opts.Tools.Unwrap() {
			ag.Tools = append(ag.Tools, Tool{
				Type: ToolType(id),
			})
		}
	}

	if err := p.client().UpdateAgent(string(p.id), UpdateAgentRequest{
		Instructions:   &ag.Instructions,
		Model:          &ag.Model,
		Name:           &ag.Name,
		Description:    &ag.Description,
		Tools:          ag.Tools,
		CompletionArgs: &ag.CompletionArgs,
	}); err != nil {
		return agent.Agent{}, err
	}

	ag, err = p.client().GetAgent(string(p.id))
	if err != nil {
		return agent.Agent{}, err
	}

	return ag.IntoAgent(), nil
}

func (p *mistralAgent) client() *Client {
	return p.parent.client()
}
