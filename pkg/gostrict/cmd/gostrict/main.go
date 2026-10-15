// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Command gostrict checks, translates, generates and runs a package of the gostrict subset.
//
//	gostrict check [-json] <dir>      reports the diagnostics, exit code 1 if there are any
//	gostrict fix <dir>                applies the mechanical fixes of the diagnostics to the files and reports
//	                                  the remaining diagnostics, exit code 1 if there are any
//	gostrict build <dir>              prints the JavaScript bundle
//	gostrict gen <dir>                writes the functions of the package as gostrict.Funcs into gostrict_gen.go
//	                                  and, for a header package, the TypeScript declaration of its host
//	gostrict run <dir> <func> [args]  runs an exported function in goja, the arguments are parsed by its kinds
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: gostrict check [-json] <dir> | fix <dir> | build <dir> | gen <dir> | run <dir> <func> [args]")
	}

	switch os.Args[1] {
	case "check":
		fs := flag.NewFlagSet("check", flag.ExitOnError)
		asJSON := fs.Bool("json", false, "print the diagnostics as JSON")
		_ = fs.Parse(os.Args[2:])
		prog, err := gostrict.Load(fs.Arg(0))
		if err != nil {
			return err
		}
		diags := gostrict.Check(prog)
		switch {
		case *asJSON:
			os.Stdout.Write(diags.JSON())
			fmt.Println()
		case len(diags) == 0:
			fmt.Println("ok")
		default:
			fmt.Print(diags.Text())
		}
		if len(diags) > 0 {
			os.Exit(1)
		}
		return nil
	case "fix":
		diags, err := fix(os.Args[2])
		if err != nil {
			return err
		}
		if len(diags) == 0 {
			fmt.Println("ok")
			return nil
		}
		fmt.Print(diags.Text())
		os.Exit(1)
		return nil
	case "gen":
		dir := os.Args[2]
		prog, err := gostrict.Load(dir)
		if err != nil {
			return err
		}
		if diags := gostrict.Check(prog); len(diags) > 0 {
			fmt.Print(diags.Text())
			os.Exit(1)
		}
		src, err := gostrict.Generate(prog)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, gostrict.GeneratedFile), src, 0o644); err != nil {
			return err
		}
		if prog.Main.Kind == gostrict.KindHeader {
			ts, err := gostrict.GenerateTS(prog)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(dir, prog.Main.Name+".host.d.ts"), ts, 0o644)
		}
		return nil
	case "build", "run":
		b, diags, err := gostrict.Compile(os.Args[2])
		if err != nil {
			return err
		}
		if len(diags) > 0 {
			fmt.Print(diags.Text())
			os.Exit(1)
		}
		if os.Args[1] == "build" {
			fmt.Println(b.JS)
			return nil
		}
		if len(os.Args) < 4 {
			return fmt.Errorf("run needs the name of an exported function")
		}
		sig, ok := b.Exports[os.Args[3]]
		if !ok {
			return fmt.Errorf("%s is not an exported function with values, which cross the boundary", os.Args[3])
		}
		args, err := parseArgs(sig, os.Args[4:])
		if err != nil {
			return err
		}

		ctx := context.Background()
		in, err := new(gojaengine.Engine).Load(ctx, b, nil)
		if err != nil {
			return err
		}
		defer in.Close()
		res, err := in.Call(ctx, sig.Name, args...)
		if err != nil {
			return err
		}
		for _, r := range res {
			fmt.Println(r)
		}
		return nil
	}
	return fmt.Errorf("unknown command %s", os.Args[1])
}

func parseArgs(sig gostrict.Func, args []string) ([]any, error) {
	if len(args) != len(sig.Params) {
		return nil, fmt.Errorf("%s takes %d arguments", sig.Name, len(sig.Params))
	}
	res := make([]any, len(args))
	for i, t := range sig.Params {
		var err error
		switch t.Kind {
		case gostrict.ValueBool:
			res[i], err = strconv.ParseBool(args[i])
		case gostrict.ValueInt32:
			var v int64
			v, err = strconv.ParseInt(args[i], 10, 32)
			res[i] = int32(v)
		case gostrict.ValueUint32:
			var v uint64
			v, err = strconv.ParseUint(args[i], 10, 32)
			res[i] = uint32(v)
		case gostrict.ValueFloat64:
			res[i], err = strconv.ParseFloat(args[i], 64)
		case gostrict.ValueString:
			res[i] = args[i]
		default:
			return nil, fmt.Errorf("argument %d: %s can not be given on the command line", i, t)
		}
		if err != nil {
			return nil, fmt.Errorf("argument %d: %w", i, err)
		}
	}
	return res, nil
}

// fix applies the edits of the diagnostics until there are none left, since a fix may enable another one.
func fix(dir string) (gostrict.Diagnostics, error) {
	for range 5 {
		prog, err := gostrict.Load(dir)
		if err != nil {
			return nil, err
		}
		diags := gostrict.Check(prog)
		files := map[string]bool{}
		for _, d := range diags {
			for _, e := range d.Edits {
				files[e.File] = true
			}
		}
		if len(files) == 0 {
			return diags, nil
		}
		for file := range files {
			src, err := os.ReadFile(file)
			if err != nil {
				return nil, err
			}
			fixed, err := gostrict.ApplyEdits(src, diags.Edits(file))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", file, err)
			}
			if err := os.WriteFile(file, fixed, 0o644); err != nil {
				return nil, err
			}
			fmt.Println("fixed", file)
		}
	}
	prog, err := gostrict.Load(dir)
	if err != nil {
		return nil, err
	}
	return gostrict.Check(prog), nil
}
