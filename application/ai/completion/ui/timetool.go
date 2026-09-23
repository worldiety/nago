// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"slices"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/presentation/core"
)

// CurrentTimeToolName is the name under which [CurrentTimeTool] is advertised.
const CurrentTimeToolName = "current_time"

// CurrentTime is what [CurrentTimeTool] returns to the model.
type CurrentTime struct {
	Now       string `json:"now"`
	Date      string `json:"date"`
	Time      string `json:"time"`
	Weekday   string `json:"weekday"`
	ISOWeek   int    `json:"isoWeek"`
	Timezone  string `json:"timezone"`
	UTCOffset string `json:"utcOffset"`
}

type currentTimeIn struct{}

// CurrentTimeTool returns a tool telling the model the current date and time in the time zone of wnd.
//
// A model has no clock. Its notion of "today" is its training cut-off, and it calculates with that date
// fluently and wrongly: "overdue since yesterday", "next Monday", "in two weeks". The time is offered as a tool
// rather than written into the system prompt on purpose - a prompt that changes every second defeats the
// prompt caching of every provider, while a tool costs only when it is actually needed.
//
// The time zone is [core.Window.Location].
//
// [ChatOptions] wires this tool automatically unless [ChatOptions.DisableCurrentTime] is set.
func CurrentTimeTool(wnd core.Window) completion.Tool {
	return currentTimeTool(wnd.Location, time.Now)
}

func currentTimeTool(location func() *time.Location, now func() time.Time) completion.Tool {
	return completion.NewTool(CurrentTimeToolName,
		"Returns the current date and time of the user, including weekday, ISO week and time zone. "+
			"You do not know the current date on your own. Call this before answering anything that depends on it, "+
			"such as today, tomorrow, next week, deadlines, ages, durations or whether something is overdue.",
		func(currentTimeIn) (CurrentTime, error) {
			loc := location()
			if loc == nil {
				loc = time.Local
			}

			t := now().In(loc)
			_, week := t.ISOWeek()

			return CurrentTime{
				Now:       t.Format(time.RFC3339),
				Date:      t.Format(time.DateOnly),
				Time:      t.Format(time.TimeOnly),
				Weekday:   t.Weekday().String(),
				ISOWeek:   week,
				Timezone:  loc.String(),
				UTCOffset: t.Format("-07:00"),
			}, nil
		})
}

// withBuiltinTool appends tool unless a tool of the same name is already present. An application which brings
// its own variant wins, and [completion.Run] does not fail on a duplicate name.
func withBuiltinTool(tools []completion.Tool, tool completion.Tool) []completion.Tool {
	if slices.ContainsFunc(tools, func(t completion.Tool) bool { return t.Def.Name == tool.Def.Name }) {
		return tools
	}

	return append(tools, tool)
}
