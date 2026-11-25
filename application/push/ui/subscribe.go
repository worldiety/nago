// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package uipush lets the user subscribe the browser to the push notifications of the application. The browser asks
// for the permission only by [Subscribe], which must be the reaction to an action of the user, e.g. the click on a
// button, which the application offers for that purpose. [Card] and [Toggle] are such views.
package uipush

import (
	"errors"
	"fmt"

	"go.wdy.de/nago/application/push"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

var (
	// ErrNotEnabled is returned, if the push module has not been enabled by cfgpush.Enable.
	ErrNotEnabled = errors.New("push notifications are not enabled")

	// ErrNoAnswer is returned by [Subscribe], if the push service of the browser has not answered in time after the
	// user has allowed the notifications, e.g. because the push service of the device is not available.
	ErrNoAnswer = errors.New("the push service of the browser has not answered")
)

// Permission is the permission of the browser for notifications.
type Permission string

const (
	// PermissionDefault means that the user has not decided yet. The browser asks at the next subscription.
	PermissionDefault Permission = "default"
	PermissionGranted Permission = "granted"
	// PermissionDenied means that the user has blocked the notifications. Only the user can allow them again in
	// the settings of the browser.
	PermissionDenied Permission = "denied"
)

// State describes the push notifications of the browser of a window.
type State struct {
	// Supported is false, if the browser cannot receive push notifications at all.
	Supported bool
	// RequiresInstall is true, if the browser receives push notifications only as an installed web app, like Safari
	// on iOS and iPadOS, and the application is not installed.
	RequiresInstall bool
	// Standalone is true, if the application runs as an installed web app.
	Standalone bool
	Permission Permission
	// Endpoint is the endpoint of the subscription of the browser, if there is one.
	Endpoint string
	// Registered is true, if the browser receives the notifications of the user of the window.
	Registered bool
}

// CanSubscribe reports, whether a [Subscribe] may succeed.
func (s State) CanSubscribe() bool {
	return s.Supported && !s.RequiresInstall && s.Permission != PermissionDenied
}

// useCases returns the use cases of the module, which cfgpush.Enable has registered.
func useCases(wnd core.Window) (push.UseCases, bool) {
	return core.FromContext[push.UseCases](wnd.Context(), "")
}

// Enabled reports, whether the push module has been enabled.
func Enabled(wnd core.Window) bool {
	_, ok := useCases(wnd)
	return ok
}

// ReadState reads the state of the browser of the window. It never asks the user and never registers a service
// worker. The callback is invoked at most once, and not at all, if the window is gone before the browser answered.
func ReadState(wnd core.Window, fn func(State, error)) {
	uc, ok := useCases(wnd)
	if !ok {
		fn(State{}, ErrNotEnabled)
		return
	}

	core.AsyncCall(wnd, &proto.CallPushState{Keep: true}, func(ret proto.CallRet) {
		switch ret := ret.(type) {
		case *proto.RetPushState:
			fn(stateOf(wnd, uc, ret), nil)
		case *proto.RetError:
			fn(State{}, errorOf(ret))
		default:
			fn(State{}, errors.New("unexpected answer of the browser"))
		}
	})
}

// Subscribe asks the browser for the permission for notifications, if the user has not decided yet, and registers
// the browser for the notifications of the user of the window. Call it only as the reaction to an action of the
// user, which the application offers to get notifications, because the browser asks only within such a user
// gesture, and because the question must not surprise the user. It requires a signed-in user and asks nothing
// otherwise. The callback is invoked at most once.
func Subscribe(wnd core.Window, fn func(State, error)) {
	uc, ok := useCases(wnd)
	if !ok {
		fn(State{}, ErrNotEnabled)
		return
	}

	// the permission would be of no use without a user to register the browser for
	if !wnd.Subject().Valid() {
		fn(State{}, user.InvalidSubjectErr)
		return
	}

	call := &proto.CallPushSubscribe{
		ApplicationServerKey: proto.Str(uc.PublicKey()),
		ServiceWorker:        proto.URI(push.ServiceWorkerPath),
	}

	core.AsyncCall(wnd, call, func(ret proto.CallRet) {
		switch ret := ret.(type) {
		case *proto.RetPushSubscription:
			_, err := uc.Register(wnd.Subject(), sessionID(wnd), push.BrowserSubscription{
				Endpoint:  string(ret.Endpoint),
				P256dh:    string(ret.P256dh),
				Auth:      string(ret.Auth),
				Origin:    string(ret.Origin),
				UserAgent: string(ret.UserAgent),
			})

			state := State{Supported: true, Permission: PermissionGranted, Endpoint: string(ret.Endpoint), Registered: err == nil}
			fn(state, err)
		case *proto.RetPushState:
			fn(stateOf(wnd, uc, ret), nil)
		case *proto.RetError:
			fn(State{}, errorOf(ret))
		default:
			fn(State{}, errors.New("unexpected answer of the browser"))
		}
	})
}

// Unsubscribe removes the subscription of the browser of the window, so that it gets no notifications anymore. The
// permission of the browser stays. The callback is invoked at most once with the new state.
func Unsubscribe(wnd core.Window, fn func(State, error)) {
	uc, ok := useCases(wnd)
	if !ok {
		fn(State{}, ErrNotEnabled)
		return
	}

	core.AsyncCall(wnd, &proto.CallPushUnsubscribe{Keep: true}, func(ret proto.CallRet) {
		switch ret := ret.(type) {
		case *proto.RetPushSubscription:
			if err := uc.Unregister(wnd.Subject(), string(ret.Endpoint)); err != nil && wnd.Subject().Valid() {
				fn(State{}, err)
				return
			}
		case *proto.RetVoid:
		case *proto.RetError:
			fn(State{}, errorOf(ret))
			return
		}

		ReadState(wnd, fn)
	})
}

func stateOf(wnd core.Window, uc push.UseCases, ret *proto.RetPushState) State {
	s := State{
		Supported:       bool(ret.Supported),
		RequiresInstall: bool(ret.RequiresInstall),
		Standalone:      bool(ret.Standalone),
		Permission:      Permission(ret.Permission),
		Endpoint:        string(ret.Endpoint),
	}

	if s.Endpoint != "" && s.Permission == PermissionGranted {
		s.Registered, _ = uc.IsRegistered(wnd.Subject(), s.Endpoint)
	}

	return s
}

func sessionID(wnd core.Window) session.ID {
	if s := wnd.Session(); s != nil {
		return s.ID()
	}

	return ""
}

// codeNoAnswer is the code of a [proto.RetError], by which the frontend tells, that the push service has not
// answered in time, see [proto.CallPushSubscribe].
const codeNoAnswer = 504

func errorOf(ret *proto.RetError) error {
	if ret.Code == codeNoAnswer {
		return fmt.Errorf("%w: %s", ErrNoAnswer, ret.Message)
	}

	return core.AsyncError{Code: int(ret.Code), Message: string(ret.Message)}
}
