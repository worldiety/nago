// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/pkg/xsync"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/markdown"
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

// agenticChat is an interactive playground for [completion.Run]: it drives the full agentic loop (call model
// -> execute requested tools -> feed results back -> repeat) and renders both the final answer and the full
// message trace so the tool calls become visible.
func agenticChat(wnd core.Window, uc ai.UseCases) core.View {
	var prov provider.Provider
	var comps completion.Completions

	for p, err := range uc.FindAllProvider(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		if c := p.Completions(); c.IsSome() {
			prov = p
			comps = c.Unwrap()
			break
		}
	}

	if comps == nil {
		return alert.BannerError(fmt.Errorf("kein Provider mit stateless Completions gefunden – bitte ein Secret konfigurieren"))
	}

	prompt := core.AutoState[string](wnd).Init(func() string {
		return "Was ist die Quadratwurzel aus (3 mal 12)? Nutze die Tools."
	})
	answer := core.AutoState[string](wnd)
	trace := core.AutoState[string](wnd)
	usage := core.AutoState[string](wnd)
	busy := core.AutoState[bool](wnd)
	selectedModel := core.AutoState[model.ID](wnd).Init(func() model.ID {
		for m, err := range comps.Models(wnd.Subject()) {
			if err != nil {
				return ""
			}
			return m.ID
		}
		return ""
	})

	submit := func() {
		question := strings.TrimSpace(prompt.Get())
		if question == "" || busy.Get() {
			return
		}

		busy.Set(true)
		answer.Set("")
		trace.Set("")
		usage.Set("")

		xsync.Go(func() error {
			res, history, err := completion.Run(wnd.Subject(), comps, completion.RunOptions{
				Options: completion.Options{
					Model:     selectedModel.Get(),
					System:    "You are a helpful assistant. Use the provided tools to compute results instead of guessing.",
					MaxTokens: 1024,
					Messages: []completion.Message{
						{
							Role:    completion.User,
							Content: []completion.Content{completion.Text{Text: question}},
						},
					},
				},
				Tools: agenticTools(),
			})

			if err != nil {
				wnd.Post(func() {
					busy.Set(false)
					alert.ShowBannerError(wnd, err)
				})
				return nil
			}

			var sb strings.Builder
			for _, c := range res.Message.Content {
				if t, ok := c.(completion.Text); ok {
					sb.WriteString(t.Text)
				}
			}

			wnd.Post(func() {
				answer.Set(sb.String())
				trace.Set(renderTrace(history))
				usage.Set(renderUsage(res.Usage))
				busy.Set(false)
			})

			return nil
		}, func(err error) {
			if err != nil {
				wnd.Post(func() {
					busy.Set(false)
					alert.ShowBannerError(wnd, err)
				})
			}
		})
	}

	return ui.VStack(
		ui.Text(fmt.Sprintf("Agentic Tool-Loop – %s (%s)", prov.Name(), selectedModel.Get())).Font(ui.Title),

		ui.TextField("Deine Eingabe", prompt.Get()).
			InputValue(prompt).
			Lines(4).
			FullWidth().
			Disabled(busy.Get()),

		ui.PrimaryButton(submit).
			Title("completion.Run starten").
			Enabled(!busy.Get()),

		ui.If(busy.Get(), ui.Text("… die KI denkt nach und ruft ggf. Tools auf")),

		ui.If(answer.Get() != "", ui.VStack(
			ui.Text("Antwort").Font(ui.SubTitle),
			markdown.RichText(answer.Get()),
		).Alignment(ui.Leading).
			FullWidth().
			BackgroundColor(ui.M2).
			Border(ui.Border{}.Radius(ui.L8)).
			Padding(ui.Padding{}.All(ui.L16))),

		ui.If(trace.Get() != "", ui.VStack(
			ui.Text("Message-Trace").Font(ui.SubTitle),
			ui.CodeEditor(trace.Get()).Language("text").FullWidth(),
		).Alignment(ui.Leading).
			FullWidth().
			BackgroundColor(ui.M2).
			Border(ui.Border{}.Radius(ui.L8)).
			Padding(ui.Padding{}.All(ui.L16))),

		ui.If(usage.Get() != "", ui.VStack(
			ui.Text("Token-Usage (letzter Turn)").Font(ui.SubTitle),
			ui.CodeEditor(usage.Get()).Language("text").FullWidth(),
		).Alignment(ui.Leading).
			FullWidth().
			BackgroundColor(ui.M2).
			Border(ui.Border{}.Radius(ui.L8)).
			Padding(ui.Padding{}.All(ui.L16))),
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

// intermediate tool calls and tool results become visible.
func renderTrace(history []completion.Message) string {
	var sb strings.Builder
	for i, msg := range history {
		fmt.Fprintf(&sb, "#%d [%s]\n", i, msg.Role)
		for _, c := range msg.Content {
			switch v := c.(type) {
			case completion.Text:
				fmt.Fprintf(&sb, "  text: %s\n", v.Text)
			case completion.Thinking:
				fmt.Fprintf(&sb, "  thinking: %s\n", v.Text)
			case completion.ToolCall:
				fmt.Fprintf(&sb, "  tool_call %s(%s) -> %s\n", v.Name, v.ID, string(v.Arguments))
			case completion.ToolResult:
				var inner strings.Builder
				for _, ic := range v.Content {
					if t, ok := ic.(completion.Text); ok {
						inner.WriteString(t.Text)
					}
				}
				marker := "ok"
				if v.IsError {
					marker = "error"
				}
				fmt.Fprintf(&sb, "  tool_result %s [%s]: %s\n", v.ToolCallID, marker, inner.String())
			default:
				raw, _ := json.Marshal(v)
				fmt.Fprintf(&sb, "  %T: %s\n", v, string(raw))
			}
		}
	}
	return sb.String()
}
