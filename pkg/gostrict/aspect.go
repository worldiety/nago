// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"errors"
	"maps"
)

// Aspect is called around each call of an entry point of a whole program, see [LoadRoots]: when the host calls it
// and when another package of the program calls it, but not when its own package calls it. So the host keeps a
// frame for each call, e.g. with the rights of the callee and a transaction of its area, which the functions of
// header packages find on the stack of the frames, while the program itself never sees it.
type Aspect interface {
	// Enter is called before the entry point with its full name, e.g. "example.com/scripts/crm/read.Main". An
	// error denies the call and aborts the whole call of the instance, like [Abort].
	Enter(entry string) error
	// Exit is called, after the entry point has returned or was aborted, with its error result, which is nil, if it
	// has none, or the error, which aborted it. So the host commits or rolls back a transaction of the frame.
	Exit(entry string, err error)
}

// EnterFunc and ExitFunc are the functions of the host, which a whole program calls around an entry point, see
// [Funcs.WithAspect].
const (
	EnterFunc = "$enter"
	ExitFunc  = "$exit"
)

// WithAspect returns a copy of the functions with the functions, which call the aspect. A whole program needs
// them, so an engine does not load it without an aspect.
func (f Funcs) WithAspect(a Aspect) Funcs {
	res := maps.Clone(f)
	if res == nil {
		res = Funcs{}
	}
	res[EnterFunc] = func(args []any) []any {
		if err := a.Enter(args[0].(string)); err != nil {
			Abort(err)
		}
		return nil
	}
	res[ExitFunc] = func(args []any) []any {
		a.Exit(args[0].(string), ErrorOf(args[1]))
		return nil
	}
	return res
}

// Aborted is the panic of [Abort].
type Aborted struct {
	Err error
}

func (a *Aborted) Error() string {
	return "aborted: " + a.Err.Error()
}

func (a *Aborted) Unwrap() error {
	return a.Err
}

// Abort ends the whole call of the instance from a function of the host, e.g. because an aspect denies an entry
// point or a database fails. The program can not catch it: the engine unwinds the calls, the aspect exits each
// entry point with the error, and the call of the instance returns err itself, so that errors.Is finds it.
func Abort(err error) {
	panic(&Aborted{Err: err})
}

// AbortedErr returns the error of an abort, which an engine has caught, or false.
func AbortedErr(v any) (error, bool) {
	switch v := v.(type) {
	case *Aborted:
		return v.Err, true
	case error:
		var a *Aborted
		if errors.As(v, &a) {
			return a.Err, true
		}
	}
	return nil, false
}
