// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"slices"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
)

func TestWebResearch(t *testing.T) {
	search, fetch := completion.WebSearchTool(), completion.WebFetchTool()

	tests := []struct {
		name      string
		assistant WebResearch
		search    bool // the operator allows the search
		fetch     bool // the operator allows the fetch
		want      []completion.ServerTool
	}{
		{name: "off by default", assistant: WebResearchGlobal},
		{name: "operator allows the search", assistant: WebResearchGlobal, search: true, want: []completion.ServerTool{search}},
		{name: "operator allows the fetch", assistant: WebResearchGlobal, fetch: true, want: []completion.ServerTool{fetch}},
		{name: "operator allows both", assistant: WebResearchGlobal, search: true, fetch: true, want: []completion.ServerTool{search, fetch}},
		{name: "assistant always despite default", assistant: WebResearchAlways, want: []completion.ServerTool{search, fetch}},
		{name: "assistant always completes the operator", assistant: WebResearchAlways, fetch: true, want: []completion.ServerTool{search, fetch}},
		{name: "assistant never despite operator", assistant: WebResearchNever, search: true, fetch: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.assistant.tools(AssistantSettings{WebSearch: tt.search, WebFetch: tt.fetch})
			if !slices.EqualFunc(got, tt.want, func(a, b completion.ServerTool) bool {
				return a.Kind == b.Kind && a.MaxUses == b.MaxUses && a.MaxContentTokens == b.MaxContentTokens
			}) {
				t.Fatalf("want %+v, got %+v", tt.want, got)
			}
		})
	}
}
