// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

// agenticTools are the executable Go functions exposed to the model. The schema for the input struct is
// derived automatically by [completion.NewTool] via reflection, so the model knows how to call them.
func agenticTools() []completion.Tool {
	type calcIn struct {
		Op string  `json:"op" desc:"the arithmetic operation, one of: add, sub, mul, div"`
		A  float64 `json:"a" desc:"the left operand"`
		B  float64 `json:"b" desc:"the right operand"`
	}
	type calcOut struct {
		Result float64 `json:"result"`
	}

	calc := completion.NewTool("calculator", "performs a basic arithmetic operation on two numbers",
		func(in calcIn) (calcOut, error) {
			switch in.Op {
			case "add":
				return calcOut{Result: in.A + in.B}, nil
			case "sub":
				return calcOut{Result: in.A - in.B}, nil
			case "mul":
				return calcOut{Result: in.A * in.B}, nil
			case "div":
				if in.B == 0 {
					return calcOut{}, fmt.Errorf("division by zero")
				}
				return calcOut{Result: in.A / in.B}, nil
			default:
				return calcOut{}, fmt.Errorf("unknown operation %q", in.Op)
			}
		})

	type sqrtIn struct {
		X float64 `json:"x" desc:"a non-negative number to take the square root of"`
	}
	type sqrtOut struct {
		Result float64 `json:"result"`
	}

	sqrt := completion.NewTool("sqrt", "computes the square root of a number",
		func(in sqrtIn) (sqrtOut, error) {
			if in.X < 0 {
				return sqrtOut{}, fmt.Errorf("cannot take the square root of a negative number")
			}
			return sqrtOut{Result: math.Sqrt(in.X)}, nil
		})

	type timeIn struct{}
	type timeOut struct {
		Now string `json:"now"`
	}

	now := completion.NewTool("current_time", "returns the current server time in RFC3339 format",
		func(in timeIn) (timeOut, error) {
			return timeOut{Now: time.Now().Format(time.RFC3339)}, nil
		})

	return []completion.Tool{calc, sqrt, now}
}

// agenticChat demonstrates the generic [uicompletion.Chat] component driving a tool-using assistant. It
// configures TWO agents so the component renders an agent picker: a general assistant with the calculator/
// sqrt/time tools and a math specialist that additionally asks the user for confirmation via the built-in
// ask_user tool (enabled through ChatOptions.AskUser). History is off here, so this chat is transient.
func agenticChat(wnd core.Window, uc ai.UseCases, sessions session.UseCases) core.View {
	prov, comps, err := firstCompletionProvider(wnd.Subject(), uc, false)
	if err != nil {
		return alert.BannerError(err)
	}

	tools := func(auth.Subject) []completion.Tool { return agenticTools() }

	chat := uicompletion.Chat(wnd, uicompletion.ChatOptions{
		Sessions:    sessions,
		Completions: comps,
		Provider:    prov,
		Title:       "Agentic Tool-Loop",
		AskUser:     true,
		Agents: []uicompletion.Agent{
			{
				ID:           "assistant",
				Name:         "Allrounder",
				SystemPrompt: "You are a helpful assistant. Use the provided tools to compute results instead of guessing.",
				Tools:        tools,
			},
			{
				ID:           "math",
				Name:         "Mathe-Spezialist",
				SystemPrompt: "You are a meticulous math specialist. Use the tools for every calculation and use ask_user to confirm ambiguous inputs before computing.",
				Tools:        tools,
			},
		},
	})

	return ui.VStack(
		ui.Text("Agentic Tool-Loop (uicompletion.Chat mit Agent-Auswahl + ask_user)").Font(ui.Title),
		ui.Text("Zwei Agenten stehen zur Auswahl. Beide nutzen dieselben Rechen-Tools; der Mathe-Spezialist fragt bei Bedarf per ask_user nach."),
		chat,
	).Alignment(ui.Leading).
		Gap(ui.L16).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L16))
}

// renderUsage formats the token accounting of the final turn. The cache fields prove whether Anthropic prompt
// caching kicked in: CacheWriteTokens > 0 means a fresh prefix was stored, CacheReadTokens > 0 means the
// stable prefix (system prompt, tools and earlier conversation turns) was served from cache at ~0.1x cost.
func renderUsage(u completion.Usage) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "input_tokens:        %d\n", u.InputTokens)
	fmt.Fprintf(&sb, "output_tokens:       %d\n", u.OutputTokens)
	fmt.Fprintf(&sb, "cache_read_tokens:   %d  (aus dem Cache gelesen, ~0.1x Kosten)\n", u.CacheReadTokens)
	fmt.Fprintf(&sb, "cache_write_tokens:  %d  (neu in den Cache geschrieben, ~1.25x Kosten)\n", u.CacheWriteTokens)
	return sb.String()
}
