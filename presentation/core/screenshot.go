// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"encoding/base64"
	"errors"
	"fmt"

	"go.wdy.de/nago/pkg/std/async"
	"go.wdy.de/nago/presentation/proto"
)

// ScreenshotOptions configures [Window.Screenshot].
type ScreenshotOptions struct {
	// Selector is an optional CSS selector. When set, only the first matching element is captured. Empty
	// captures the whole page.
	Selector string

	// Image requests a rendered PNG of the current DOM.
	Image bool

	// Snapshot requests a textual accessibility snapshot of the current DOM, similar to the aria snapshot of
	// Playwright: roles, names, values and states as an indented tree. It is usually cheaper and more precise
	// for a language model than the image.
	Snapshot bool

	// MaxEdge limits the longest edge of the PNG in pixels. Zero means the frontend default of 1568, which
	// is the largest size typical vision models process without downscaling.
	MaxEdge int
}

// Screenshot is the result of [Window.Screenshot].
type Screenshot struct {
	// PNG contains the encoded image, if requested.
	PNG []byte

	// Width and Height are the pixel dimensions of PNG.
	Width, Height int

	// Snapshot contains the textual accessibility snapshot, if requested.
	Snapshot string
}

// ErrScreenshotUnavailable is returned if the window cannot be captured, e.g. because it is not backed by a
// live frontend.
var ErrScreenshotUnavailable = errors.New("screenshot unavailable")

// Screenshot asks the frontend to capture its current render state.
//
// The DOM is rendered into a PNG by the frontend itself, which is an approximation rather than a pixel
// capture: cross-origin images without CORS headers, iframe contents and videos may be missing. This is
// intended as a feedback channel for automated agents, which is also why no permission is involved - the
// developer must wire it explicitly.
//
// Pending state changes are rendered and sent before the capture is requested, so a screenshot taken right
// after a mutation shows its result.
func (s *scopeWindow) Screenshot(opts ScreenshotOptions) *async.Future[Screenshot] {
	var fut async.Future[Screenshot]

	if !opts.Image && !opts.Snapshot {
		opts.Image = true
	}

	maxEdge := opts.MaxEdge
	if maxEdge < 0 {
		maxEdge = 0
	}

	call := &proto.CallScreenshot{
		Keep:     true,
		Selector: proto.Str(opts.Selector),
		Image:    proto.Bool(opts.Image),
		Snapshot: proto.Bool(opts.Snapshot),
		MaxEdge:  proto.Uint(maxEdge),
	}

	posted := s.Post(func() {
		// Flush pending state changes first. The frontend processes messages in order, so the capture then
		// sees the DOM of the latest render and not the one before it.
		if p := s.parent; p != nil && (p.dirty || p.hasDirtyStates()) {
			p.forceRender(0)
			p.dirty = false
		}

		var cancel func()
		cancel = AsyncCall(s, call, func(ret proto.CallRet) {
			if cancel != nil {
				cancel()
			}

			switch ret := ret.(type) {
			case *proto.RetScreenshot:
				shot := Screenshot{
					Width:    int(ret.Width),
					Height:   int(ret.Height),
					Snapshot: string(ret.Snapshot),
				}

				if ret.PngBase64 != "" {
					buf, err := base64.StdEncoding.DecodeString(string(ret.PngBase64))
					if err != nil {
						fut.Set(Screenshot{}, fmt.Errorf("cannot decode screenshot: %w", err))
						return
					}

					shot.PNG = buf
				}

				fut.Set(shot, nil)
			case *proto.RetError:
				fut.Set(Screenshot{}, newAsyncError(ret))
			default:
				fut.Set(Screenshot{}, fmt.Errorf("unexpected screenshot result: %T", ret))
			}
		})
	})

	if !posted {
		fut.Set(Screenshot{}, ErrScreenshotUnavailable)
	}

	return &fut
}
