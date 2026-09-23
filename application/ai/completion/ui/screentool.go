// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

// DefaultScreenToolTimeout bounds how long [ScreenTool] waits for the frontend to answer.
const DefaultScreenToolTimeout = 30 * time.Second

// maxScreenSnapshotRunes caps the accessibility snapshot handed to the model, so a huge table cannot exhaust the
// context window on its own.
const maxScreenSnapshotRunes = 60_000

// ScreenToolOptions configures [ScreenTool].
type ScreenToolOptions struct {
	// Name of the tool. Defaults to inspect_screen.
	Name string

	// Timeout for the frontend to answer. Zero means [DefaultScreenToolTimeout].
	Timeout time.Duration

	// MaxEdge limits the longest edge of the PNG in pixels. Zero means the frontend default.
	MaxEdge int

	// DisableImage removes the image option, e.g. for models without vision support. The tool then only
	// returns the textual snapshot.
	DisableImage bool
}

type screenToolIn struct {
	Image    bool   `json:"image" optional:"true" desc:"when true, additionally return a PNG of the rendered screen. Costs considerably more tokens than the text snapshot; request it only for questions about visual appearance such as layout, colors, overlaps or cut-off content"`
	Selector string `json:"selector" optional:"true" desc:"optional CSS selector restricting the inspection to the first matching element; leave empty for the whole screen"`
}

type screenToolInNoImage struct {
	Selector string `json:"selector" optional:"true" desc:"optional CSS selector restricting the inspection to the first matching element; leave empty for the whole screen"`
}

// ScreenTool returns a ready-made tool that lets the model look at the screen of wnd, like an embedded
// Playwright: it always returns an accessibility snapshot of what is currently rendered (roles, names, values
// and states as an indented tree) and, on request, a PNG of it.
//
// It is meant as the feedback channel for an assistant which changes the UI on behalf of the user: after a
// mutation the model can verify what the user actually sees, instead of assuming that its action had the
// intended effect. Pending state changes are rendered before the capture, see [core.Window.Screenshot].
//
// Unlike most tools it is bound to a window, so build it where the window is known, e.g. inside the decorator
// together with the agent:
//
//	Tools: append(tools, uicompletion.ScreenTool(wnd, uicompletion.ScreenToolOptions{})),
//
// There is no permission involved: the model only sees what is rendered for the acting user anyway, and the
// developer has to wire the tool explicitly. The image is sent inline, so it stays part of the conversation
// history and is billed again on every following turn - which is why the model is told to prefer the text.
func ScreenTool(wnd core.Window, opts ScreenToolOptions) completion.Tool {
	name := opts.Name
	if name == "" {
		name = "inspect_screen"
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultScreenToolTimeout
	}

	desc := "Inspects what the user currently sees in the application window and returns an accessibility snapshot " +
		"of the rendered screen: an indented tree of roles (heading, button, textbox, checkbox, …) with their names, " +
		"current values and states such as [checked] or [disabled]. " +
		"Use it after you changed something to verify the result instead of assuming it worked, " +
		"and whenever the user refers to something on the screen. " +
		"The snapshot also contains the assistant chat itself; ignore that part."

	if opts.DisableImage {
		return completion.NewSubjectContentTool(name, desc, func(_ auth.Subject, in screenToolInNoImage) ([]completion.Content, error) {
			return inspectScreen(wnd, screenToolIn{Selector: in.Selector}, opts.MaxEdge, timeout)
		})
	}

	desc += " Set image to true to additionally receive a PNG, but only when the visual appearance matters."

	return completion.NewSubjectContentTool(name, desc, func(_ auth.Subject, in screenToolIn) ([]completion.Content, error) {
		return inspectScreen(wnd, in, opts.MaxEdge, timeout)
	})
}

func inspectScreen(wnd core.Window, in screenToolIn, maxEdge int, timeout time.Duration) ([]completion.Content, error) {
	shot, err := AwaitScreenshot(wnd, core.ScreenshotOptions{
		Selector: strings.TrimSpace(in.Selector),
		Image:    in.Image,
		Snapshot: true,
		MaxEdge:  maxEdge,
	}, timeout)
	if err != nil {
		return nil, fmt.Errorf("cannot inspect the screen: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Route: %s\n", routeOf(wnd))
	if in.Selector != "" {
		fmt.Fprintf(&sb, "Selector: %s\n", in.Selector)
	}

	snapshot := shot.Snapshot
	truncated := false
	if r := []rune(snapshot); len(r) > maxScreenSnapshotRunes {
		snapshot = string(r[:maxScreenSnapshotRunes])
		truncated = true
	}

	if strings.TrimSpace(snapshot) == "" {
		sb.WriteString("Accessibility snapshot: (empty - nothing accessible is rendered)\n")
	} else {
		sb.WriteString("Accessibility snapshot:\n")
		sb.WriteString(snapshot)
		sb.WriteString("\n")
	}

	if truncated {
		fmt.Fprintf(&sb, "[truncated after %d characters; use a selector to inspect a part of the screen]\n", maxScreenSnapshotRunes)
	}

	content := []completion.Content{}

	if len(shot.PNG) > 0 {
		fmt.Fprintf(&sb, "A PNG of %dx%d pixels follows. It is re-rendered from the DOM, so external images, iframes and videos may be missing.\n", shot.Width, shot.Height)
	}

	content = append(content, completion.Text{Text: sb.String()})

	if len(shot.PNG) > 0 {
		content = append(content, completion.Media{MimeType: file.PNG, Source: completion.Source{Data: shot.PNG}})
	}

	return content, nil
}

// ErrScreenshotTimeout is returned by [AwaitScreenshot] if the frontend did not answer in time.
var ErrScreenshotTimeout = errors.New("the frontend did not answer in time")

// AwaitScreenshot calls [core.Window.Screenshot] and blocks until the frontend answered or the timeout elapsed.
// It must not be called from the event loop of wnd, because the capture is dispatched through it; call it from
// a background goroutine such as a tool invocation.
func AwaitScreenshot(wnd core.Window, opts core.ScreenshotOptions, timeout time.Duration) (core.Screenshot, error) {
	type result struct {
		shot core.Screenshot
		err  error
	}

	ch := make(chan result, 1)
	closeObs := wnd.Screenshot(opts).Observe(func(shot core.Screenshot, err error) {
		select {
		case ch <- result{shot, err}:
		default:
		}
	})
	defer closeObs()

	select {
	case r := <-ch:
		return r.shot, r.err
	case <-time.After(timeout):
		return core.Screenshot{}, ErrScreenshotTimeout
	}
}

func routeOf(wnd core.Window) core.NavigationPath {
	if p := wnd.Path(); p != "" {
		return p
	}

	return "."
}
