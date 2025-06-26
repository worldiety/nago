// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"go.wdy.de/nago/presentation/proto"
	"strconv"
)

// AsyncCall requests an asynchronous invocation within the frontend implementation. Usually, this is used
// to trigger some native blocking behavior or to query some specific (hardware) information. The lifecycle
// is identical to conventional function callbacks, which means that the callback is removed automatically just
// right before the next render cycle starts. It is also never guaranteed, that a result will ever occur, either due to
// the lifecyle or because of a communication interruption or because the user never confirms something required to
// continue at the frontend-side.
//
// To know which calls are defined and how they respond, you have to inspect the according [proto.CallArgs]
// documentation of the concrete implementing types.
//
// Also note that by convention, the raw protocol types should never be used by application directly. Instead,
// an application developer should always prefer the correctly typed wrappers which may be scattered across the
// types of this package where they make sense.
func AsyncCall(wnd Window, args proto.CallArgs, fn func(ret proto.CallRet)) (cancel func()) {
	w := wnd.(*scopeWindow)
	ptr := proto.Ptr(w.lastAsyncInvokePtr.Add(1))
	w.asyncCallbacks.Put(ptr, fn)

	w.parent.Publish(&proto.CallRequested{
		CallPtr: ptr,
		Call:    args,
	})

	return func() {
		w.asyncCallbacks.Delete(ptr)
	}
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
