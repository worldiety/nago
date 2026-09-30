// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"context"
	"iter"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

type fakeProvider struct {
	id     provider.ID
	models []model.ID
}

func (p fakeProvider) Identity() provider.ID             { return p.id }
func (p fakeProvider) Name() string                      { return string(p.id) }
func (p fakeProvider) Description() string               { return "" }
func (p fakeProvider) Files() option.Opt[provider.Files] { return option.None[provider.Files]() }
func (p fakeProvider) Completions() option.Opt[completion.Completions] {
	return option.Some[completion.Completions](p)
}

func (p fakeProvider) Models(auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) {
		for _, m := range p.models {
			if !yield(model.Model{ID: m}, nil) {
				return
			}
		}
	}
}

func (p fakeProvider) Complete(context.Context, auth.Subject, completion.Options) (completion.Result, error) {
	return completion.Result{}, nil
}

func (p fakeProvider) Stream(context.Context, auth.Subject, completion.Options) iter.Seq2[completion.Delta, error] {
	return nil
}

func TestResolve(t *testing.T) {
	anthropic := fakeProvider{id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1", models: []model.ID{"claude-sonnet", "claude-haiku"}}
	openai := fakeProvider{id: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb1", models: []model.ID{"gpt-5", "openai/gpt-4o"}}

	a := &Assistant{useCases: ai.UseCases{FindAllProvider: func(auth.Subject) iter.Seq2[provider.Provider, error] {
		return func(yield func(provider.Provider, error) bool) {
			_ = yield(anthropic, nil) && yield(openai, nil)
		}
	}}}

	tests := []struct {
		name         string
		choice       ModelChoice
		wantProvider provider.ID
		wantModel    model.ID
	}{
		{name: "no choice takes the first provider and model", wantProvider: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1", wantModel: "claude-sonnet"},
		{name: "choice of the second provider", choice: NewModelChoice("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb1", "gpt-5"), wantProvider: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb1", wantModel: "gpt-5"},
		{name: "model id with a slash", choice: NewModelChoice("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb1", "openai/gpt-4o"), wantProvider: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb1", wantModel: "openai/gpt-4o"},
		{name: "legacy bare model id of the second provider", choice: "gpt-5", wantProvider: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb1", wantModel: "gpt-5"},
		{name: "legacy bare model id with a slash", choice: "openai/gpt-4o", wantProvider: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb1", wantModel: "openai/gpt-4o"},
		{name: "legacy bare model id is never replaced", choice: "claude-sonnet-alias", wantProvider: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1", wantModel: "claude-sonnet-alias"},
		{name: "deleted provider falls back to the default", choice: NewModelChoice("ccccccccccccccccccccccccccccccc1", "gpt-5"), wantProvider: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1", wantModel: "claude-sonnet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, m, err := a.Resolve(user.SU(), AssistantSettings{Model: tt.choice})
			if err != nil {
				t.Fatal(err)
			}

			if c.Provider.Identity() != tt.wantProvider || m != tt.wantModel {
				t.Fatalf("want %s/%s, got %s/%s", tt.wantProvider, tt.wantModel, c.Provider.Identity(), m)
			}
		})
	}
}
