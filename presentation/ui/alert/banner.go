// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package alert

import (
	"strings"
	"time"

	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
)

const (
	IconOK      = "<svg width=\"20\" height=\"20\" viewBox=\"0 0 20 20\" fill=\"none\" xmlns=\"http://www.w3.org/2000/svg\">\n<path fill-rule=\"evenodd\" clip-rule=\"evenodd\" d=\"M1.66675 10.0001C1.66675 5.39771 5.39771 1.66675 10.0001 1.66675C14.6025 1.66675 18.3334 5.39771 18.3334 10.0001C18.3334 14.6025 14.6025 18.3334 10.0001 18.3334C5.39771 18.3334 1.66675 14.6025 1.66675 10.0001ZM9.50666 5.41675C9.04643 5.41675 8.67333 5.78984 8.67333 6.25008C8.67333 6.71032 9.04643 7.08341 9.50666 7.08341H9.515C9.97523 7.08341 10.3483 6.71032 10.3483 6.25008C10.3483 5.78984 9.97523 5.41675 9.515 5.41675H9.50666ZM8.33341 8.33341C7.87318 8.33341 7.50008 8.70651 7.50008 9.16675C7.50008 9.62698 7.87318 10.0001 8.33341 10.0001H9.16675V12.5001H8.33341C7.87318 12.5001 7.50008 12.8732 7.50008 13.3334C7.50008 13.7937 7.87318 14.1667 8.33341 14.1667H11.6667C12.127 14.1667 12.5001 13.7937 12.5001 13.3334C12.5001 12.8732 12.127 12.5001 11.6667 12.5001H10.8334V9.16675C10.8334 8.70651 10.4603 8.33341 10.0001 8.33341H8.33341Z\"/>\n</svg>\n"
	IconError   = "<svg width=\"18\" height=\"16\" viewBox=\"0 0 18 16\" fill=\"none\" xmlns=\"http://www.w3.org/2000/svg\">\n<path d=\"M15.9132 15.4053C15.88 15.4053 15.8429 15.4023 15.797 15.3975H1.52651C1.48159 15.4023 1.4435 15.4053 1.41225 15.4053C0.947411 15.4053 0.51284 15.1855 0.250145 14.8174C-0.079933 14.3545 -0.0838393 13.7568 0.241356 13.2939L7.4728 0.732422C7.7355 0.274414 8.18081 0 8.66323 0C9.1437 0 9.58804 0.274414 9.85073 0.732422L17.0812 13.2939C17.4054 13.7559 17.4044 14.3535 17.0783 14.8174C16.8136 15.1855 16.3781 15.4053 15.9132 15.4053ZM8.58022 11.5781C8.57827 11.5791 8.57436 11.5791 8.57436 11.5791C8.21401 11.582 7.87417 11.7646 7.66421 12.0674C7.51186 12.2822 7.43081 12.5498 7.43667 12.8223C7.4562 13.4766 7.96304 13.9961 8.59097 14.0039H8.60171H8.62417C8.98647 13.9932 9.32534 13.8096 9.53139 13.5127C9.68472 13.292 9.76479 13.0225 9.756 12.7529C9.73354 12.1113 9.23647 11.5957 8.62417 11.5791H8.59389C8.59194 11.5781 8.58999 11.5781 8.58706 11.5781H8.58022ZM8.59682 3.24121C8.04604 3.24121 7.5978 3.68945 7.5978 4.24023V9.01562C7.5978 9.56738 8.04604 10.0166 8.59682 10.0166C9.14761 10.0166 9.59487 9.56738 9.59487 9.01562V4.24023C9.59487 3.68945 9.14761 3.24121 8.59682 3.24121Z\"/>\n</svg>\n"
	IconWarning = "<svg width=\"20\" height=\"20\" viewBox=\"0 0 20 20\" fill=\"none\" xmlns=\"http://www.w3.org/2000/svg\">\n<path fill-rule=\"evenodd\" clip-rule=\"evenodd\" d=\"M0.5 10L10 0.5L19.5 10L10 19.5L0.5 10ZM9.50658 5.41667C9.04634 5.41667 8.67325 5.78976 8.67325 6.25C8.67325 6.71024 9.04634 7.08333 9.50658 7.08333H9.51492C9.97515 7.08333 10.3482 6.71024 10.3482 6.25C10.3482 5.78976 9.97515 5.41667 9.51492 5.41667H9.50658ZM8.33333 8.33333C7.8731 8.33333 7.5 8.70643 7.5 9.16667C7.5 9.6269 7.8731 10 8.33333 10H9.16667V12.5H8.33333C7.8731 12.5 7.5 12.8731 7.5 13.3333C7.5 13.7936 7.8731 14.1667 8.33333 14.1667H11.6667C12.1269 14.1667 12.5 13.7936 12.5 13.3333C12.5 12.8731 12.1269 12.5 11.6667 12.5H10.8333V9.16667C10.8333 8.70643 10.4602 8.33333 10 8.33333H8.33333Z\"/>\n</svg>\n"
	IconSuccess = "<svg width=\"20\" height=\"20\" viewBox=\"0 0 20 20\" fill=\"none\" xmlns=\"http://www.w3.org/2000/svg\">\n<path fill-rule=\"evenodd\" clip-rule=\"evenodd\" d=\"M10.0001 1.66675C9.34066 1.66675 8.70823 1.92865 8.24188 2.39486L7.49762 3.13912C7.34388 3.29291 7.13519 3.37934 6.91775 3.37944H5.86629C5.20673 3.37944 4.57419 3.64144 4.10782 4.10782C3.64144 4.57419 3.37944 5.20673 3.37944 5.86629V6.91792C3.37934 7.13537 3.29282 7.34397 3.13904 7.49771L2.39486 8.24188C1.92865 8.70823 1.66675 9.34066 1.66675 10.0001C1.66675 10.6595 1.92874 11.292 2.39495 11.7584L3.13912 12.5025C3.29291 12.6563 3.37934 12.865 3.37944 13.0824V14.1339C3.37944 14.7934 3.64144 15.426 4.10782 15.8923C4.57419 16.3587 5.20673 16.6207 5.86629 16.6207H6.91792C7.13537 16.6208 7.34397 16.7073 7.49771 16.8611L8.24188 17.6053C8.70823 18.0715 9.34066 18.3334 10.0001 18.3334C10.6595 18.3334 11.292 18.0714 11.7584 17.6052L12.5025 16.861C12.6563 16.7073 12.865 16.6208 13.0824 16.6207H14.1339C14.7934 16.6207 15.426 16.3587 15.8923 15.8923C16.3587 15.426 16.6207 14.7934 16.6207 14.1339V13.0822C16.6208 12.8648 16.7073 12.6562 16.8611 12.5025L17.6053 11.7583C18.0715 11.2919 18.3334 10.6595 18.3334 10.0001C18.3334 9.34066 18.0714 8.70814 17.6052 8.24179L16.861 7.49762C16.7073 7.34388 16.6208 7.13519 16.6207 6.91775V5.86629C16.6207 5.20673 16.3587 4.57419 15.8923 4.10782C15.426 3.64144 14.7934 3.37944 14.1339 3.37944H13.0822C12.8648 3.37934 12.6562 3.29282 12.5025 3.13904L11.7583 2.39486C11.2919 1.92865 10.6595 1.66675 10.0001 1.66675ZM13.0696 8.10907C13.3951 7.78363 13.3951 7.256 13.0696 6.93056C12.7442 6.60512 12.2165 6.60512 11.8911 6.93056L8.34656 10.4751L7.2823 9.41083C6.95686 9.08539 6.42923 9.08539 6.10379 9.41083C5.77835 9.73627 5.77835 10.2639 6.10379 10.5893L7.75731 12.2429C8.08274 12.5683 8.61038 12.5683 8.93582 12.2429L13.0696 8.10907Z\"/>\n</svg>\n"
)

// TBanner is a feedback component(Banner).
// This component displays a prominent message to the user,
// typically used for notifications, warnings, or confirmations. It consists
// of a title and message, and can optionally be dismissible and styled
// according to intent (e.g., success, warning, error).
//
// It also supports a callback when the banner is closed.
type TBanner struct {
	id               string
	title            string
	message          string
	presented        *core.State[bool]
	onClosed         func()
	intent           Intent
	frame            ui.Frame
	autoCloseTimeout time.Duration // Deprecated: Will be removed because of accessibility concerns.
}

func Banner(title, message string) TBanner {
	return TBanner{
		title:   title,
		message: message,
		frame:   ui.Frame{Width: ui.L400, MaxWidth: ui.Full},
	}
}

// ID sets an optional ID for the banner component
func (t TBanner) ID(id string) TBanner {
	t.id = id
	return t
}

// AutoCloseTimeoutOrDefault either takes the given duration d or timeouts after 5 seconds.
// Deprecated: Will be removed because of accessibility concerns.
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
		ico = core.SVG(IconOK)
	case IntentWarning:
		textColor = ui.ColorBannerWarningText
		bgColor = ui.ColorBannerWarningBackground
		ico = core.SVG(IconWarning)
	case IntentSuccess:
		textColor = ui.ColorBannerSuccessText
		bgColor = ui.ColorBannerSuccessBackground
		ico = core.SVG(IconSuccess)

	default:
		ico = core.SVG(IconError)
		textColor = ui.ColorBannerErrorText
		bgColor = ui.ColorBannerErrorBackground
	}

	return ui.VStack(
		ui.HStack(
			ui.Image().
				FillColor(textColor).
				Embed(ico).
				Frame(ui.Frame{}.Size(ui.L24, ui.L24)),
			ui.Text(strings.ToUpper(t.title)).
				Font(ui.Font{
					Size:       "1.25rem",
					Weight:     700,
					LineHeight: "1.25rem",
				}).
				Color(textColor).
				Padding(ui.Padding{Top: ui.L2}),
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
			}).
				Padding(ui.Padding{}.All(ui.L4)),
			),
		).Gap(ui.L12).
			FullWidth(),
		ui.Text(t.message).Color(textColor),
		ui.IfFunc(t.intent == IntentOk && t.presented != nil && t.autoCloseTimeout > 0, func() core.View {
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
		ID(t.id).
		Border(ui.Border{}.Radius(ui.L12)).
		Padding(ui.Padding{}.All(ui.L20)).
		Frame(t.frame).Render(ctx)
}
