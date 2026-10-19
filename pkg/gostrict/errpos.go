// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"errors"
	"strconv"
)

// Position is a line of the Go source of a program, e.g. example.com/app/main.go:12.
type Position struct {
	// File is the import path of the package and the name of the file, like in the source map.
	File string
	Line int
}

func (p Position) String() string {
	return p.File + ":" + strconv.Itoa(p.Line)
}

// ProgramError is an error, which a program has created, e.g. with errors.New, and an engine has returned, with
// the position in the Go source, where the program created it. Its message is that of the program, so it compares
// like any other error in [SameValues].
type ProgramError struct {
	Msg  string
	File string
	Line int
}

func (e *ProgramError) Error() string {
	return e.Msg
}

// ProgramError returns the error of the program with the message and the index of its position in
// [Bundle.Positions], or a plain error, if the index is invalid, e.g. for an error, which a binding of a header
// package created. An engine converts the errors of the program with it.
func (b *Bundle) ProgramError(msg string, pos int) error {
	if pos < 0 || pos >= len(b.Positions) {
		return errors.New(msg)
	}
	p := b.Positions[pos]
	return &ProgramError{Msg: msg, File: p.File, Line: p.Line}
}

// ErrorPos returns the position in the Go source, where the program created the error, which an engine returned,
// e.g. from an exported function or to [Aspect.Exit]. Only errors.New and the functions of the library, which
// return an error, e.g. text.ParseInt, record their position, and only engines, which run the translated code,
// e.g. gojaengine, report it. The native engine runs the Go code itself and returns its errors unchanged, so it
// reports none, nor does the browser, nor does a host function or an error type of the program.
func ErrorPos(err error) (file string, line int, ok bool) {
	var pe *ProgramError
	if errors.As(err, &pe) {
		return pe.File, pe.Line, true
	}
	return "", 0, false
}
