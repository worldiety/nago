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
	// error denies the call and aborts the whole call of the instance, like [Abort], without an Exit of the denied
	// entry point. An error, which [AsResult] marks, denies the call instead with a value: the entry point returns
	// zero values, whose lists and maps are empty, and the unmarked error as its results, without its body and
	// without Exit, so that its caller sees the error like one, which the entry point returned itself. An entry
	// point, whose last result is not an error, aborts with the unmarked error.
	Enter(entry string) error
	// Exit is called, after the entry point has returned or was aborted, with its error result, which is nil, if it
	// has none, or the error, which aborted it. So the host commits or rolls back a transaction of the frame.
	Exit(entry string, err error)
}

// EnterFunc and ExitFunc are the functions of the host, which a whole program calls around an entry point, see
// [Funcs.WithAspect]. EnterFunc takes the full name of the entry point and whether its last result is an error,
// and returns the error, which the entry point returns instead of running, or nil.
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
		err := a.Enter(args[0].(string))
		if err == nil {
			return []any{nil}
		}
		if r, ok := err.(*ResultError); ok {
			if args[1].(bool) {
				return []any{r.Err}
			}
			err = r.Err
		}
		Abort(err)
		return nil
	}
	res[ExitFunc] = func(args []any) []any {
		a.Exit(args[0].(string), ErrorOf(args[1]))
		return nil
	}
	return res
}

// ResultError is the mark of an error of [Aspect.Enter], which the denied entry point returns as its error result,
// see [AsResult], and the panic of [ReturnError].
type ResultError struct {
	Err error
}

func (e *ResultError) Error() string {
	return e.Err.Error()
}

func (e *ResultError) Unwrap() error {
	return e.Err
}

// AsResult marks err, which [Aspect.Enter] returns, so that the denied entry point returns zero values and err as
// its results instead of aborting the whole call, e.g. a denial, which the calling package may handle. Only the mark
// of the error, which Enter returns itself, counts, not one, which another error wraps. AsResult(nil) is nil.
func AsResult(err error) error {
	if err == nil {
		return nil
	}
	return &ResultError{Err: err}
}

// ReturnError ends the entry point, in which a function of the host runs, from that function with zero values, whose
// lists and maps are empty, and err as its results, e.g. because the generated code of the entry point rejects its
// input: the engine unwinds the calls within the entry point, the aspect exits it with err, and its caller handles
// err like an error, which the entry point returned. Within an entry point, whose last result is not an error, or
// outside of any entry point, it aborts the whole call with err like [Abort]. Only a function, which the program
// calls itself, may end its entry point, not one within a callback, which the host calls.
func ReturnError(err error) {
	panic(&ResultError{Err: err})
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

// AbortedErr returns the error of an abort, which an engine has caught, or false. A [ReturnError], which no entry
// point returned, is an abort as well.
func AbortedErr(v any) (error, bool) {
	switch v := v.(type) {
	case *Aborted:
		return v.Err, true
	case *ResultError:
		return v.Err, true
	case error:
		var a *Aborted
		if errors.As(v, &a) {
			return a.Err, true
		}
	}
	return nil, false
}
