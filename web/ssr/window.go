// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ssr

import (
	"context"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

// ssrNavigation is a no-op implementation of core.Navigation used during SSR.
type ssrNavigation struct{}

func (n ssrNavigation) ForwardTo(_ core.NavigationPath, _ core.Values)                 {}
func (n ssrNavigation) ForwardToTarget(_ core.NavigationPath, _ string, _ core.Values) {}
func (n ssrNavigation) BackwardTo(_ core.NavigationPath, _ core.Values)                {}
func (n ssrNavigation) Back()                                                          {}
func (n ssrNavigation) ResetTo(_ core.NavigationPath, _ core.Values)                   {}
func (n ssrNavigation) Reload()                                                        {}
func (n ssrNavigation) Open(_ core.URI)                                                {}

// ssrSession is an empty stub for session.UserSession used during SSR.
type ssrSession struct{}

func (s ssrSession) ID() session.ID                         { return "ssr" }
func (s ssrSession) User() std.Option[user.ID]              { return std.None[user.ID]() }
func (s ssrSession) CreatedAt() std.Option[time.Time]       { return std.None[time.Time]() }
func (s ssrSession) AuthenticatedAt() std.Option[time.Time] { return std.None[time.Time]() }
func (s ssrSession) PutString(_ string, _ string) error     { return nil }
func (s ssrSession) GetString(_ string) (string, bool)      { return "", false }

// ssrClipboard is a no-op clipboard for SSR.
type ssrClipboard struct{}

func (c ssrClipboard) SetText(_ string) error            { return nil }
func (c ssrClipboard) Text(onResult func(string, error)) { onResult("", nil) }

// ssrWindow implements core.Window as a no-op stub for server-side rendering.
// It provides:
//   - SizeClass2XL + Light color scheme
//   - Locale from Accept-Language header (parsed by caller)
//   - Subject from user.GetAnonUser
//   - All interactive methods as no-ops
type ssrWindow struct {
	locale      language.Tag
	bundle      *i18n.Bundle
	subject     user.Subject
	path        core.NavigationPath
	application *core.Application
}

func (w *ssrWindow) AddInputListener(elemID string, fn func(evt core.InputEvent)) (close func()) {
	return func() {

	}
}

// NewWindow creates a new SSR window with the given locale and anonymous subject.
func NewWindow(locale language.Tag, getAnonUser user.GetAnonUser, path core.NavigationPath) core.Window {
	bundle, ok := i18n.Default.MatchBundle(locale)
	if !ok {
		bundle, _ = i18n.Default.MatchBundle(language.English)
	}
	return &ssrWindow{
		locale:  locale,
		bundle:  bundle,
		subject: getAnonUser(),
		path:    path,
	}
}

func (w *ssrWindow) Navigation() core.Navigation  { return ssrNavigation{} }
func (w *ssrWindow) Values() core.Values          { return core.Values{} }
func (w *ssrWindow) Subject() user.Subject        { return w.subject }
func (w *ssrWindow) UpdateSubject(_ user.Subject) {}
func (w *ssrWindow) Context() context.Context     { return w.subject.Context() }
func (w *ssrWindow) Session() session.UserSession { return ssrSession{} }
func (w *ssrWindow) Locale() language.Tag         { return w.locale }
func (w *ssrWindow) Location() *time.Location     { return time.UTC }
func (w *ssrWindow) Info() core.WindowInfo {
	return core.WindowInfo{
		Width:       1920,
		Height:      1080,
		Density:     1,
		SizeClass:   core.SizeClass2XL,
		ColorScheme: core.Light,
	}
}
func (w *ssrWindow) ExportFiles(_ core.ExportFilesOptions) {}
func (w *ssrWindow) ImportFiles(_ core.ImportFilesOptions) {}
func (w *ssrWindow) SetColorScheme(_ core.ColorScheme)     {}
func (w *ssrWindow) Application() *core.Application        { return w.application }
func (w *ssrWindow) Path() core.NavigationPath             { return w.path }
func (w *ssrWindow) AddDestroyObserver(_ func()) func()    { return func() {} }
func (w *ssrWindow) Clipboard() core.Clipboard             { return ssrClipboard{} }
func (w *ssrWindow) Logout() error                         { return nil }
func (w *ssrWindow) MediaDevices() core.MediaDevices       { return core.MediaDevices{} }
func (w *ssrWindow) Bundle() *i18n.Bundle                  { return w.bundle }
func (w *ssrWindow) Post(_ func()) bool                    { return false }
func (w *ssrWindow) PostDelayed(_ func(), _ time.Duration) {}
func (w *ssrWindow) RequestFocus(_ string)                 {}
