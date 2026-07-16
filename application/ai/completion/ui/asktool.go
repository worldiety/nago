// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"strings"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/markdown"
)

// pendingAsk holds a clarifying question the model asked back to the user via the ask_user tool. reply is a
// buffered channel the (background) tool goroutine blocks on until the user answered in the UI.
type pendingAsk struct {
	Question string
	Options  []string
	reply    chan string
}

// askUserTool builds the built-in ask_user clarification tool. When the model calls it, the run's background
// goroutine posts a [pendingAsk] into the ask state (marshalled back onto the UI event loop via wnd.Post) and
// blocks until the user answers via [renderAsk]. It is wired automatically when [ChatOptions.AskUser] is set.
func askUserTool(wnd core.Window, ask *core.State[*pendingAsk]) completion.Tool {
	type askIn struct {
		Question string   `json:"question" desc:"the clarifying question to ask the user"`
		Options  []string `json:"options,omitempty" desc:"optional predefined answers the user may pick from"`
	}
	type askOut struct {
		Answer string `json:"answer"`
	}

	return completion.NewTool("ask_user",
		"asks the user a clarifying question and waits for their answer before continuing. Use this whenever you need a decision or missing information from the user.",
		func(in askIn) (askOut, error) {
			ch := make(chan string, 1)
			wnd.Post(func() {
				ask.Set(&pendingAsk{Question: in.Question, Options: in.Options, reply: ch})
			})
			answer := <-ch
			return askOut{Answer: answer}, nil
		})
}

// answerAsk delivers the user's answer to the blocked tool goroutine and clears the pending question.
func answerAsk(ask *core.State[*pendingAsk], pa *pendingAsk, answer string) {
	if pa == nil {
		return
	}
	select {
	case pa.reply <- answer:
	default:
	}
	ask.Set(nil)
}

// renderAsk renders a clarifying question the model asked, offering the predefined options as buttons plus a
// free-text fallback.
func renderAsk(wnd core.Window, ask *core.State[*pendingAsk], pa *pendingAsk) core.View {
	free := core.AutoState[string](wnd)

	var optionViews []core.View
	for _, opt := range pa.Options {
		opt := opt
		optionViews = append(optionViews, ui.HStack(ui.Text(opt)).Action(func() {
			answerAsk(ask, pa, opt)
		}).FullWidth().Border(ui.Border{}.Color(ui.I0).Width(ui.L1).Radius(ui.L8)).Padding(ui.Padding{}.All(ui.L8)))
	}

	return ui.VStack(
		ui.Text("Rückfrage der KI").Font(ui.TitleSmall),
		markdown.RichText(pa.Question),
		ui.VStack(optionViews...).Gap(ui.L8).FullWidth().Alignment(ui.Leading),
		ui.TextField("Eigene Antwort", free.Get()).
			InputValue(free).
			FullWidth().
			KeydownEnter(func() {
				if strings.TrimSpace(free.Get()) != "" {
					answerAsk(ask, pa, free.Get())
				}
			}),
		ui.HStack(
			ui.Spacer(),
			ui.PrimaryButton(func() {
				answerAsk(ask, pa, free.Get())
			}).Title("Antworten").Enabled(strings.TrimSpace(free.Get()) != ""),
		).FullWidth(),
	).Gap(ui.L8).FullWidth().Alignment(ui.Leading).
		BackgroundColor(ui.M2).
		Border(ui.Border{}.Radius(ui.L8)).
		Padding(ui.Padding{}.All(ui.L8))
}
