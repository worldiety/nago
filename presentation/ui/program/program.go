// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package program runs programs of the gostrict subset in the frontend, see package pkg/gostrict, e.g. the scene
// of a three.js component or the drawing of a 2D canvas, which must react within a frame and can not wait for
// the server. The frontend loads the translated program once per version from [BundlePath], calls its exported
// functions and exchanges messages with its backend in Go through a [Channel]. Input events of the component go
// to the program directly as records of go.wdy.de/nago/pkg/gostrict/lib/input.
package program

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"path"
	"runtime/debug"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// cache compiles the programs of the process once per version of their sources.
var cache = gostrict.NewCache(256)

// Compile checks and translates the package main of the modules once per version of its sources. A program, which
// a user or a language model has written, must come from a module, which is not trusted, see
// [gostrict.Module.Trusted].
func Compile(main string, modules ...gostrict.Module) (*gostrict.Bundle, gostrict.Diagnostics, error) {
	return cache.Compile(main, modules...)
}

// MustCompile is like [Compile], but panics, if the program can not be translated, e.g. for a program, which an
// application embeds.
func MustCompile(main string, modules ...gostrict.Module) *gostrict.Bundle {
	b, diags, err := Compile(main, modules...)
	if err != nil {
		panic(fmt.Errorf("cannot compile %s: %w", main, err))
	}
	if len(diags) > 0 {
		panic(fmt.Errorf("cannot compile %s:\n%s", main, diags.Text()))
	}
	return b
}

// CompileFS compiles the program in the directory dir of a tree of sources, which the application embeds, e.g.
//
//	//go:embed scene
//	var sources embed.FS
//
//	bundle := program.MustCompileFS(sources, "scene")
//
// The module path is derived from the imports of the program, e.g. of its package of messages, or, if it imports
// none of the tree, from the main package, which is right, if the main package embeds the sources. The module path
// must be the import path of the sources in Go, because the backend names the messages by the types of Go. A
// tree, which matches neither or has its own header packages, needs [Compile] with an explicit [gostrict.Module].
func CompileFS(fsys fs.FS, dir string) (*gostrict.Bundle, gostrict.Diagnostics, error) {
	module, ok, err := gostrict.InferModulePath(fsys)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		info, found := debug.ReadBuildInfo()
		if !found || info.Path == "" || info.Path == "command-line-arguments" {
			return nil, nil, fmt.Errorf("cannot derive the module path of %s, use Compile with an explicit module", dir)
		}
		module = info.Path
	}
	return Compile(path.Join(module, dir), gostrict.Module{Path: module, FS: fsys})
}

// MustCompileFS is like [CompileFS], but panics, if the program can not be translated.
func MustCompileFS(fsys fs.FS, dir string) *gostrict.Bundle {
	b, diags, err := CompileFS(fsys, dir)
	if err != nil {
		panic(fmt.Errorf("cannot compile %s: %w", dir, err))
	}
	if len(diags) > 0 {
		panic(fmt.Errorf("cannot compile %s:\n%s", dir, diags.Text()))
	}
	return b
}

// Spec describes the program, which a component runs. The components embed it and offer setters for it.
type Spec struct {
	Bundle *gostrict.Bundle
	// Init is the exported function, which the component calls first, e.g. "NewGame". If it returns a state, the
	// component passes the state to the other functions, which take one.
	Init string
	// Args are the arguments of Init as values of Go, e.g. int32 or a record.
	Args []any
	// Frame is the exported function, which the component calls in each animation frame with the time in seconds.
	Frame string
	// Resize is the exported function, which the component calls with the aspect ratio, with the width and the
	// height or with the width, the height and the device pixel ratio of its area, whenever it changes.
	Resize string
	// Inputs are the input events, which the component passes to the handler of input.Event of the program.
	Inputs []core.InputEventType
	// Receive handles the messages of the program, usually the Dispatch of a [gostrict.Router], see
	// [Channel.OnReceive].
	Receive func(ctx context.Context, e gostrict.Envelope) error
}

// Render registers the bundle, connects the channel of the component and returns its description for the
// protocol. Mistakes, like a missing export, are logged, so that the view still renders.
func (s Spec) Render(wnd core.Window, id string) proto.GostrictProgram {
	if s.Bundle == nil {
		slog.Error("the component has no program", "id", id)
		return proto.GostrictProgram{}
	}

	res := proto.GostrictProgram{
		Bundle: proto.URI(Register(s.Bundle)),
		Init:   proto.Str(s.Init),
		Frame:  proto.Str(s.Frame),
		Resize: proto.Str(s.Resize),
	}
	for _, name := range []string{s.Init, s.Frame, s.Resize} {
		if _, ok := s.Bundle.Exports[name]; name != "" && !ok {
			slog.Error("the program does not export the function", "id", id, "func", name)
		}
	}
	if f, ok := s.Bundle.Exports[s.Init]; ok {
		if args, err := initArgs(s.Bundle, f, s.Args); err != nil {
			slog.Error("invalid arguments of the program", "id", id, "err", err)
		} else {
			res.Args = proto.Str(args)
		}
	}
	for _, t := range s.Inputs {
		res.Inputs = append(res.Inputs, proto.InputEventType(t))
	}

	if hasMessages(s.Bundle) {
		ch := ChannelOf(wnd, id)
		ch.attach(s.Bundle)
		if s.Receive != nil {
			ch.OnReceive(s.Receive)
		}
	}
	return res
}

// hasMessages tells, whether the program handles or sends messages.
func hasMessages(b *gostrict.Bundle) bool {
	if len(b.Handlers) > 0 {
		return true
	}
	for _, f := range b.Exports {
		for _, r := range f.Results {
			if r.Kind == gostrict.ValueOut {
				return true
			}
		}
	}
	return false
}

// initArgs converts the arguments of Init into JSON.
func initArgs(b *gostrict.Bundle, f gostrict.Func, args []any) (string, error) {
	neutral := make([]any, len(args))
	for i, a := range args {
		n, err := gostrict.ToNeutral(a)
		if err != nil {
			return "", err
		}
		neutral[i] = n
	}
	if err := gostrict.CheckArgs(f.Name, f.Params, neutral); err != nil {
		return "", err
	}
	return codec{b: b}.encode(neutral, f.Params)
}
