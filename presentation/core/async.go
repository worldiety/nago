// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"fmt"
	"log/slog"
	"strconv"

	"go.wdy.de/nago/presentation/proto"
)

// AsyncCall requests an asynchronous invocation within the frontend implementation. Usually, this is used
// to trigger some native blocking behavior or to query some specific (hardware) information. The callback is
// invoked at most once and removed automatically after the frontend has resolved the call. Use the returned cancel
// function to drop the callback earlier. It is also never guaranteed, that a result will ever occur, either due to
// the lifecycle or because of a communication interruption or because the user never confirms something required to
// continue at the frontend-side.
//
// To know which calls are defined and how they respond, you have to inspect the according [proto.CallArgs]
// documentation of the concrete implementing types.
//
// Also note that by convention, the raw protocol types should never be used by application directly. Instead,
// an application developer should always prefer the correctly typed wrappers which may be scattered across the
// types of this package where they make sense.
func AsyncCall(wnd Window, args proto.CallArgs, fn func(ret proto.CallRet)) (cancel func()) {
	return asyncCall(wnd, args, fn, false)
}

// asyncCall is like [AsyncCall] but if repeatable is true, the callback survives any resolution and is only
// removed by the returned cancel function. This is required for frontend calls which resolve multiple times,
// like input event listeners.
func asyncCall(wnd Window, args proto.CallArgs, fn func(ret proto.CallRet), repeatable bool) (cancel func()) {
	w, ok := wnd.(*scopeWindow)
	if !ok {
		slog.Error("AsyncCall: invalid window type", "wnd", fmt.Errorf("%T", wnd))
		return func() {}
	}
	ptr := w.parent.ids.nextAsync()
	if fn != nil {
		w.asyncCallbacks.Put(ptr, asyncCallback{fn: fn, repeatable: repeatable})
	}

	w.parent.Publish(&proto.CallRequested{
		CallPtr: ptr,
		Call:    args,
	})

	return func() {
		w.asyncCallbacks.Delete(ptr)
	}
}

type asyncCallback struct {
	fn         func(ret proto.CallRet)
	repeatable bool
}

type AsyncError struct {
	Code    int
	Message string
}

func newAsyncError(err *proto.RetError) AsyncError {
	return AsyncError{
		Code:    int(err.Code),
		Message: string(err.Message),
	}
}

func (e AsyncError) Error() string {
	return e.Message + " (" + strconv.Itoa(e.Code) + ")"
}
