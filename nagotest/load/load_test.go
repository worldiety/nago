// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package load_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/nagotest/load"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func configure(cfg *application.Configurator) {
	cfg.SetApplicationID("de.worldiety.nagotest")
	cfg.RootView("counter", func(wnd core.Window) core.View {
		count := core.AutoState[int](wnd)
		name := core.AutoState[string](wnd)
		greeting := core.AutoState[string](wnd)
		loaded := core.AutoState[string](wnd)

		core.OnAppear(wnd, "", func(ctx context.Context) {
			loaded.Set("loaded")
		})

		return ui.VStack(
			ui.Text(fmt.Sprintf("count: %d", count.Get())),
			ui.PrimaryButton(func() { count.Set(count.Get() + 1) }).Title("increment"),
			ui.TextField("Name", name.Get()).InputValue(name).KeydownEnter(func() {
				greeting.Set("hello " + name.Get())
			}),
			ui.Text(greeting.Get()),
			ui.Text(loaded.Get()),
		)
	})
}

func TestRun(t *testing.T) {
	url := nagotest.Serve(t, configure)

	report, err := load.Run(context.Background(), load.Config{
		URL:        url,
		Users:      5,
		RampUp:     50 * time.Millisecond,
		Iterations: 3,
		Thresholds: load.Thresholds{MaxErrorRate: 0.5},
	}, load.Scenario{
		Name: "increment",
		Path: "counter",
		Run: func(u *load.User, w *nagotest.Window) {
			// clicks on a tree which is replaced by the render of OnAppear in the meantime are discarded
			w.WaitFor(nagotest.Text("loaded"), 5*time.Second)
			w.Click(w.Find(nagotest.Text("increment")))
			w.Find(nagotest.Text("count: 1"))
		},
	}, load.Scenario{
		Name: "broken",
		Path: "counter",
		Run: func(u *load.User, w *nagotest.Window) {
			w.Find(nagotest.Text("does not exist"))
			t.Error("a failed assertion must abort the iteration")
		},
	})
	if !errors.Is(err, load.ErrThresholds) {
		t.Fatalf("expected threshold violation, got %v", err)
	}

	total, failed := report.Iterations()
	if total != 15 || failed == 0 || failed == total {
		t.Fatalf("unexpected iterations: total %d, failed %d", total, failed)
	}

	if len(report.Errors) != 1 || !strings.Contains(report.Errors[0].Message, "does not exist") {
		t.Fatalf("unexpected errors %v", report.Errors)
	}

	var buf strings.Builder
	if err := report.WriteText(&buf); err != nil {
		t.Fatal(err)
	}
	t.Log("\n" + buf.String())

	for _, want := range []string{"increment", "open counter", "click Text(\"increment\")", "THRESHOLD VIOLATIONS"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("report misses %q", want)
		}
	}
}

func TestFile(t *testing.T) {
	url := nagotest.Serve(t, configure)

	f, err := load.LoadFile("testdata/counter.json")
	if err != nil {
		t.Fatal(err)
	}

	cfg := f.Config()
	cfg.URL = url
	report, err := load.Run(context.Background(), cfg, f.ScenarioList()...)

	var buf strings.Builder
	_ = report.WriteText(&buf)
	t.Log("\n" + buf.String())

	if err != nil {
		t.Fatal(err)
	}

	if total, failed := report.Iterations(); total != 12 || failed != 0 {
		t.Fatalf("unexpected iterations: total %d, failed %d", total, failed)
	}
}

func TestLoadFileRejectsInvalidSteps(t *testing.T) {
	for name, content := range map[string]string{
		"unknown field": `{"scenarios":[{"name":"a","path":"b","steps":[{"clik":{"text":"x"}}]}]}`,
		"two actions":   `{"scenarios":[{"name":"a","path":"b","steps":[{"click":{"text":"x"},"reload":true}]}]}`,
		"no action":     `{"scenarios":[{"name":"a","path":"b","steps":[{}]}]}`,
		"input type":    `{"scenarios":[{"name":"a","path":"b","steps":[{"input":{"id":"c","type":"tap"}}]}]}`,
		"duration":      `{"duration":5,"scenarios":[{"name":"a","path":"b","steps":[]}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "f.json")
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}

			if _, err := load.LoadFile(path); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
