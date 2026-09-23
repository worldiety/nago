// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"encoding/json"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/completion"
)

func TestCurrentTimeTool_ReportsTimeInWindowLocation(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Skip("no tzdata")
	}

	// 2026-09-23 22:30 UTC is already Thursday in Berlin - the case a UTC clock gets wrong.
	fixed := time.Date(2026, 9, 23, 22, 30, 5, 0, time.UTC)
	tool := currentTimeTool(func() *time.Location { return berlin }, func() time.Time { return fixed })

	if tool.Def.Name != "current_time" || tool.Mutating {
		t.Fatalf("unexpected tool %#v", tool.Def)
	}

	raw, err := tool.Invoke(nil, json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	var got CurrentTime
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}

	want := CurrentTime{
		Now:       "2026-09-24T00:30:05+02:00",
		Date:      "2026-09-24",
		Time:      "00:30:05",
		Weekday:   "Thursday",
		ISOWeek:   39,
		Timezone:  "Europe/Berlin",
		UTCOffset: "+02:00",
	}

	if got != want {
		t.Fatalf("got %#v\nwant %#v", got, want)
	}
}

func TestCurrentTimeTool_AcceptsNoArguments(t *testing.T) {
	tool := currentTimeTool(func() *time.Location { return nil }, time.Now)
	if _, err := tool.Invoke(nil, nil); err != nil {
		t.Fatalf("nil location or empty args must not fail: %v", err)
	}
}

func TestWithBuiltinTool_ApplicationToolWins(t *testing.T) {
	own := completion.NewTool("current_time", "the application's own clock", func(struct{}) (string, error) { return "own", nil })
	builtin := currentTimeTool(func() *time.Location { return time.UTC }, time.Now)

	tools := withBuiltinTool([]completion.Tool{own}, builtin)
	if len(tools) != 1 || tools[0].Def.Description != own.Def.Description {
		t.Fatalf("builtin must not shadow or duplicate the application tool: %#v", tools)
	}

	tools = withBuiltinTool(nil, builtin)
	if len(tools) != 1 || tools[0].Def.Name != "current_time" {
		t.Fatalf("builtin not appended: %#v", tools)
	}
}
