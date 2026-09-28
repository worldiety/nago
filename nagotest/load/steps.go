// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package load

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
)

// A File is a declarative load test, usually loaded from JSON by [LoadFile]. Each scenario is a list of
// steps, which map 1:1 to the [nagotest.Window] API. All strings of selectors and values support the
// placeholders ${user}, ${iteration}, ${random} and ${data.<key>} (see [Config.Data]).
//
//	{
//	  "url": "http://localhost:3000",
//	  "users": 50, "rampUp": "30s", "duration": "5m", "thinkTime": "1s",
//	  "thresholds": {"maxErrorRate": 0.01, "p95": "500ms"},
//	  "data": [{"mail": "a@example.com", "password": "secret"}],
//	  "scenarios": [{
//	    "name": "counter", "path": "counter",
//	    "steps": [
//	      {"type": {"label": "Name"}, "value": "user ${user}"},
//	      {"click": {"text": "increment"}},
//	      {"expect": {"text": "count: 1"}},
//	      {"waitFor": {"textContains": "loaded"}, "timeout": "5s"},
//	      {"think": "2s"}
//	    ]
//	  }]
//	}
type File struct {
	URL        string              `json:"url"`
	Users      int                 `json:"users"`
	RampUp     Duration            `json:"rampUp"`
	Duration   Duration            `json:"duration"`
	Iterations int                 `json:"iterations"`
	ThinkTime  Duration            `json:"thinkTime"`
	Seed       uint64              `json:"seed"`
	Thresholds FileThresholds      `json:"thresholds"`
	Data       []map[string]string `json:"data"`
	Scenarios  []FileScenario      `json:"scenarios"`
}

// FileThresholds see [Thresholds].
type FileThresholds struct {
	MaxErrorRate float64  `json:"maxErrorRate"`
	P95          Duration `json:"p95"`
}

// FileScenario see [Scenario].
type FileScenario struct {
	Name   string            `json:"name"`
	Weight int               `json:"weight"`
	Path   string            `json:"path"`
	Values map[string]string `json:"values"`
	Steps  []Step            `json:"steps"`
}

// A Step performs exactly one action.
type Step struct {
	// Click clicks the selected node, see [nagotest.Window.Click].
	Click *Selector `json:"click,omitempty"`
	// Type sets Value into the selected input, see [nagotest.Window.Type].
	Type  *Selector `json:"type,omitempty"`
	Value string    `json:"value,omitempty"`
	// Enter presses enter within the selected field, see [nagotest.Window.PressEnter].
	Enter *Selector `json:"enter,omitempty"`
	// Expect asserts, that the selection is found. If Count is set, exactly Count nodes are expected.
	Expect *Selector `json:"expect,omitempty"`
	Count  *int      `json:"count,omitempty"`
	// Absent asserts, that nothing is selected.
	Absent *Selector `json:"absent,omitempty"`
	// WaitFor waits up to Timeout (default 10s) until the selection is found, see [nagotest.Window.WaitFor].
	WaitFor *Selector `json:"waitFor,omitempty"`
	Timeout Duration  `json:"timeout,omitempty"`
	// Input delivers an input event, see [nagotest.Window.Input].
	Input *InputStep `json:"input,omitempty"`
	// Reload reloads the window, see [nagotest.Window.Reload].
	Reload bool `json:"reload,omitempty"`
	// Think pauses the user.
	Think Duration `json:"think,omitempty"`
}

// Selector selects nodes. All given criteria must match.
type Selector struct {
	Text         string `json:"text,omitempty"`
	TextContains string `json:"textContains,omitempty"`
	Label        string `json:"label,omitempty"`
	ID           string `json:"id,omitempty"`
	// Type is the name of the protocol type, e.g. TextField.
	Type string `json:"type,omitempty"`
	// Within restricts the search to the subtree of exactly one other selection.
	Within *Selector `json:"within,omitempty"`
	// Index picks the n-th of multiple matches, starting at 0.
	Index *int `json:"index,omitempty"`
}

// InputStep see [core.InputEvent].
type InputStep struct {
	ID string `json:"id"`
	// Type is one of pointerDown, pointerUp, pointerMove, pointerCancel, keyDown, keyUp or invalidate.
	Type string  `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Code string  `json:"code"`
}

// Duration is a [time.Duration] which is encoded as string like "1m30s" in JSON.
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(buf []byte) error {
	var s string
	if err := json.Unmarshal(buf, &s); err != nil {
		return fmt.Errorf("duration must be a string like \"1m30s\": %w", err)
	}

	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(v)
	return nil
}

// LoadFile reads and validates a JSON step file. Unknown fields are rejected to reveal typos.
func LoadFile(path string) (*File, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewReader(buf))
	dec.DisallowUnknownFields()

	var f File
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	if err := f.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return &f, nil
}

var inputTypes = map[string]core.InputEventType{
	"pointerDown":   core.InputEventPointerDown,
	"pointerUp":     core.InputEventPointerUp,
	"pointerMove":   core.InputEventPointerMove,
	"pointerCancel": core.InputEventPointerCancel,
	"keyDown":       core.InputEventKeyDown,
	"keyUp":         core.InputEventKeyUp,
	"invalidate":    core.InputEventInvalidate,
}

// Validate checks that each step performs exactly one action.
func (f *File) Validate() error {
	if len(f.Scenarios) == 0 {
		return fmt.Errorf("no scenarios defined")
	}

	for _, sc := range f.Scenarios {
		if sc.Name == "" {
			return fmt.Errorf("a scenario has no name")
		}

		for i, step := range sc.Steps {
			actions := 0
			for _, set := range []bool{step.Click != nil, step.Type != nil, step.Enter != nil, step.Expect != nil,
				step.Absent != nil, step.WaitFor != nil, step.Input != nil, step.Reload, step.Think > 0} {
				if set {
					actions++
				}
			}

			if actions != 1 {
				return fmt.Errorf("scenario %q step %d: expected exactly one action but found %d", sc.Name, i+1, actions)
			}

			if step.Input != nil {
				if _, ok := inputTypes[step.Input.Type]; !ok {
					return fmt.Errorf("scenario %q step %d: unknown input type %q", sc.Name, i+1, step.Input.Type)
				}
			}
		}
	}

	return nil
}

// Config returns the load profile of the file.
func (f *File) Config() Config {
	return Config{
		URL:        f.URL,
		Users:      f.Users,
		RampUp:     time.Duration(f.RampUp),
		Duration:   time.Duration(f.Duration),
		Iterations: f.Iterations,
		ThinkTime:  time.Duration(f.ThinkTime),
		Seed:       f.Seed,
		Data:       f.Data,
		Thresholds: Thresholds{MaxErrorRate: f.Thresholds.MaxErrorRate, P95: time.Duration(f.Thresholds.P95)},
	}
}

// ScenarioList converts the declared scenarios into executable ones.
func (f *File) ScenarioList() []Scenario {
	var res []Scenario
	for _, sc := range f.Scenarios {
		res = append(res, Scenario{
			Name:   sc.Name,
			Weight: sc.Weight,
			Path:   core.NavigationPath(sc.Path),
			Values: sc.Values,
			Run: func(u *User, w *nagotest.Window) {
				for _, step := range sc.Steps {
					step.run(u, w)
				}
			},
		})
	}

	return res
}

func (s Step) run(u *User, w *nagotest.Window) {
	switch {
	case s.Click != nil:
		w.Click(s.Click.one(u, w))
	case s.Type != nil:
		w.Type(s.Type.one(u, w), expand(s.Value, u))
	case s.Enter != nil:
		w.PressEnter(s.Enter.one(u, w))
	case s.Expect != nil:
		sel := s.Expect.all(u, w)
		if s.Count != nil {
			sel.Exactly(*s.Count)
		} else if sel.Len() == 0 {
			u.Fatalf("expected %s", s.Expect.matcher(u))
		}
	case s.Absent != nil:
		s.Absent.all(u, w).None()
	case s.WaitFor != nil:
		timeout := time.Duration(s.Timeout)
		if timeout <= 0 {
			timeout = nagotest.SettleTimeout
		}
		w.WaitFor(s.WaitFor.matcher(u), timeout)
	case s.Input != nil:
		w.Input(expand(s.Input.ID, u), core.InputEvent{
			Type: inputTypes[s.Input.Type],
			X:    s.Input.X,
			Y:    s.Input.Y,
			Code: expand(s.Input.Code, u),
		})
	case s.Reload:
		w.Reload()
	case s.Think > 0:
		u.Think(time.Duration(s.Think))
	}
}

func (s *Selector) matcher(u *User) nagotest.Matcher {
	var ms []nagotest.Matcher
	if s.Text != "" {
		ms = append(ms, nagotest.Text(expand(s.Text, u)))
	}
	if s.TextContains != "" {
		ms = append(ms, nagotest.TextContains(expand(s.TextContains, u)))
	}
	if s.Label != "" {
		ms = append(ms, nagotest.Label(expand(s.Label, u)))
	}
	if s.ID != "" {
		ms = append(ms, nagotest.ID(expand(s.ID, u)))
	}
	if s.Type != "" {
		ms = append(ms, nagotest.TypeName(s.Type))
	}

	if len(ms) == 1 {
		return ms[0]
	}

	return nagotest.And(ms...)
}

// all selects all matching nodes, optionally within another selection.
func (s *Selector) all(u *User, w *nagotest.Window) nagotest.Selection {
	var sel nagotest.Selection
	if s.Within != nil {
		sel = s.Within.one(u, w).FindAll(s.matcher(u))
	} else {
		sel = w.FindAll(s.matcher(u))
	}

	if s.Index != nil {
		return sel.At(*s.Index)
	}

	return sel
}

// one selects exactly one node.
func (s *Selector) one(u *User, w *nagotest.Window) nagotest.Selection {
	return s.all(u, w).Exactly(1)
}

var placeholder = regexp.MustCompile(`\$\{([a-zA-Z0-9_.-]+)\}`)

// expand replaces the placeholders ${user}, ${iteration}, ${random} and ${data.<key>}.
func expand(s string, u *User) string {
	if !strings.Contains(s, "${") {
		return s
	}

	return placeholder.ReplaceAllStringFunc(s, func(m string) string {
		name := m[2 : len(m)-1]
		switch {
		case name == "user":
			return strconv.Itoa(u.ID)
		case name == "iteration":
			return strconv.Itoa(u.Iteration)
		case name == "random":
			return strconv.Itoa(u.Rand.IntN(1_000_000))
		case strings.HasPrefix(name, "data."):
			if v, ok := u.Data[strings.TrimPrefix(name, "data.")]; ok {
				return v
			}
		}

		u.Fatalf("unknown placeholder %s", m)
		return m
	})
}
