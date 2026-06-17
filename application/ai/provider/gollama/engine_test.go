// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"strings"
	"testing"
)

// runGate feeds the pushes through a streamGate, collecting everything surfaced via emit. It returns the
// concatenated emitted text, the full accumulated text and the gate itself for state assertions.
func runGate(stops, markers []string, abortAfter int, pushes ...string) (emitted, full string, g *streamGate) {
	var sb strings.Builder
	count := 0
	g = newStreamGate(stops, markers, func(s string) bool {
		sb.WriteString(s)
		count++
		if abortAfter > 0 && count >= abortAfter {
			return false
		}
		return true
	})
	for _, p := range pushes {
		g.push(p)
	}
	full = g.finish()
	return sb.String(), full, g
}

func TestStreamGateEmitsAllWithoutStopsOrMarkers(t *testing.T) {
	emitted, full, g := runGate(nil, nil, 0, "Hello", " ", "world")
	if emitted != "Hello world" {
		t.Errorf("emitted = %q, want %q", emitted, "Hello world")
	}
	if full != "Hello world" {
		t.Errorf("full = %q, want %q", full, "Hello world")
	}
	if g.stopped || g.masked || g.aborted {
		t.Errorf("unexpected state: stopped=%v masked=%v aborted=%v", g.stopped, g.masked, g.aborted)
	}
}

func TestStreamGateTrimsStopInSinglePush(t *testing.T) {
	emitted, full, g := runGate([]string{"<|im_end|>"}, nil, 0, "Hello<|im_end|>leftover")
	if emitted != "Hello" {
		t.Errorf("emitted = %q, want %q", emitted, "Hello")
	}
	if full != "Hello" {
		t.Errorf("full = %q, want %q (stop string and trailing text must be dropped)", full, "Hello")
	}
	if !g.stopped {
		t.Error("gate should be stopped after hitting a stop string")
	}
}

func TestStreamGateTrimsStopSplitAcrossPushes(t *testing.T) {
	// The stop string straddles two pushes; the hold-back window must keep the partial marker buffered so it
	// is recognised once completed.
	emitted, full, g := runGate([]string{"<|im_end|>"}, nil, 0, "Hello<|im_", "end|>world")
	if emitted != "Hello" {
		t.Errorf("emitted = %q, want %q", emitted, "Hello")
	}
	if full != "Hello" {
		t.Errorf("full = %q, want %q", full, "Hello")
	}
	if !g.stopped {
		t.Error("gate should be stopped")
	}
}

func TestStreamGateMasksToolMarker(t *testing.T) {
	// Once a tool-call marker appears, no further text is surfaced, but the full text (including the raw
	// tool-call syntax) is retained for the parser.
	emitted, full, g := runGate([]string{"<|im_end|>"}, []string{"<tool_call>"}, 0, `Sure<tool_call>{"name":"x"}`)
	if emitted != "Sure" {
		t.Errorf("emitted = %q, want %q (raw tool-call syntax must never be surfaced)", emitted, "Sure")
	}
	if full != `Sure<tool_call>{"name":"x"}` {
		t.Errorf("full = %q, want it to retain the tool-call section", full)
	}
	if !g.masked {
		t.Error("gate should be masked after a tool marker")
	}
}

func TestStreamGateMaskStaysOffAfterMarker(t *testing.T) {
	// Text that arrives after the marker (in later pushes) must also be suppressed.
	emitted, _, _ := runGate(nil, []string{"<tool_call>"}, 0, "A<tool_call>", `{"name":"x"}`, "trailing")
	if emitted != "A" {
		t.Errorf("emitted = %q, want %q", emitted, "A")
	}
}

func TestStreamGateHoldsBackThenFlushesOnFinish(t *testing.T) {
	// With a long marker configured, short text is shorter than the hold-back window and is buffered until
	// finish proves no marker is forming.
	emitted, full, _ := runGate(nil, []string{"<tool_call>"}, 0, "Hi")
	if emitted != "Hi" {
		t.Errorf("emitted = %q, want %q (held-back text must flush on finish)", emitted, "Hi")
	}
	if full != "Hi" {
		t.Errorf("full = %q, want %q", full, "Hi")
	}
}

func TestStreamGateAbortStopsEmission(t *testing.T) {
	emitted, _, g := runGate(nil, nil, 1, "first", "second", "third")
	if !g.aborted {
		t.Error("gate should be aborted once emit returns false")
	}
	if emitted != "first" {
		t.Errorf("emitted = %q, want only %q before abort", emitted, "first")
	}
}
