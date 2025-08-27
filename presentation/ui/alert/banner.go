// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package alert

import (
	"time"

	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
)

// TBanner is a feedback component(Banner).
// This component displays a prominent message to the user,
// typically used for notifications, warnings, or confirmations. It consists
// of a title and message, and can optionally be dismissible and styled
// according to intent (e.g., success, warning, error).
//
// It also supports a callback when the banner is closed.
type TBanner struct {
	title            string
	message          string
	presented        *core.State[bool]
	onClosed         func()
	intent           Intent
	frame            ui.Frame
	autoCloseTimeout time.Duration
}

func Banner(title, message string) TBanner {
	return TBanner{
		title:            title,
		message:          message,
		frame:            ui.Frame{Width: ui.L400, MaxWidth: "calc(100% - 2rem)"},
		autoCloseTimeout: 5 * time.Second,
	}
}

// AutoCloseTimeoutOrDefault either takes the given duration d or timeouts after 5 seconds.
func (t TBanner) AutoCloseTimeoutOrDefault(d time.Duration) TBanner {
	if d <= 0 {
		d = time.Second * 5
	}

	t.autoCloseTimeout = d
	return t
}

// Frame sets a custom frame (layout constraints) for the banner.
func (t TBanner) Frame(frame ui.Frame) TBanner {
	t.frame = frame
	return t
}

// Closeable makes the banner dismissible by binding its visibility to the given state.
func (t TBanner) Closeable(presented *core.State[bool]) TBanner {
	t.presented = presented
	return t
}

// OnClosed sets a callback function that is triggered when the banner is closed.
func (t TBanner) OnClosed(fn func()) TBanner {
	t.onClosed = fn
	return t
}

// Intent sets the visual intent of the banner (e.g., success, warning, error).
func (t TBanner) Intent(intent Intent) TBanner {
	t.intent = intent
	return t
}

// Render builds and displays the banner component with styling and behavior
// based on its intent (e.g., success/info or error). It shows an icon, title,
// and message, and optionally provides a dismiss button. If auto-close is
// enabled, a countdown progress bar is displayed and the banner closes itself
// after the timeout. The banner adapts colors (text/background) according to
// its intent and applies padding, spacing, and rounded borders for consistent
// styling.
func (t TBanner) Render(ctx core.RenderContext) core.RenderNode {
	if t.presented != nil && !t.presented.Get() {
		return ui.HStack().Render(ctx)
	}

	var textColor ui.Color
	var bgColor ui.Color
	var ico core.SVG

	switch t.intent {
	case IntentOk:
		textColor = ui.ColorBannerInfoText
		bgColor = ui.ColorBannerInfoBackground
		ico = heroSolid.Check

	default:
		ico = heroSolid.ExclamationTriangle
		textColor = ui.ColorBannerErrorText
		bgColor = ui.ColorBannerErrorBackground
	}

	return ui.VStack(
		ui.HStack(
			ui.Image().
				FillColor(textColor).
				Embed(ico).
				Frame(ui.Frame{}.Size(ui.L20, ui.L20)),
			ui.Text(t.title).
				Font(ui.SubTitle).
				Color(textColor),
			ui.Spacer(),
			ui.If(t.presented != nil, ui.HStack(ui.Image().
				Embed(heroSolid.XMark).
				FillColor(textColor).
				Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
			).Action(func() {
				t.presented.Set(false)
				if t.onClosed != nil {
					t.onClosed()
				}
			})),
		).Gap(ui.L4).
			FullWidth(),
		ui.Text(t.message).Color(textColor),
		ui.IfFunc(t.intent == IntentOk && t.presented != nil, func() core.View {
			targetTime := core.DerivedState[time.Time](t.presented, "ctt").Init(func() time.Time {
				return time.Now().Add(t.autoCloseTimeout)
			})

			// TODO something is fishy here
			// TODO is this a problem of recycling function pointer ids between allocated views?
			duration := targetTime.Get().Sub(time.Now())
			//fmt.Println(duration, targetTime.ID())
			if duration < 0 {
				t.presented.Set(false)
				t.presented.Invalidate()
				if t.onClosed != nil {
					t.onClosed()
				}
			}

			duration = max(duration, 0)

			return ui.CountDown(duration).
				Done(!t.presented.Get()).
				Style(ui.CountDownStyleProgress).
				ProgressColor(ui.ColorBannerInfoText).
				Frame(ui.Frame{}.FullWidth()).
				Action(func() {
					t.presented.Set(false)
					t.presented.Invalidate()
					//fmt.Println("set presented to false", t.presented.ID())
					if t.onClosed != nil {
						t.onClosed()
					}
				})
		}),
	).Alignment(ui.Leading).
		Gap(ui.L8).
		BackgroundColor(bgColor).
		Border(ui.Border{}.Radius(ui.L12)).
		Padding(ui.Padding{}.All(ui.L16)).
		Frame(t.frame).Render(ctx)
}
