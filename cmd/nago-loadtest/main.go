// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// nago-loadtest puts a running nago server under load, as declared by a JSON step file.
// See [load.File] for the file format and example.json for an example.
//
//	nago-loadtest -f example.json -url http://localhost:3000 -users 100 -duration 5m -report report.json
//
// The exit code is 1 if the thresholds of the file are violated and 2 on any other error.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.wdy.de/nago/nagotest/load"
)

func main() {
	os.Exit(run())
}

func run() int {
	file := flag.String("f", "", "the JSON step file (required)")
	url := flag.String("url", "", "overrides the url of the file")
	users := flag.Int("users", 0, "overrides the amount of concurrent users")
	duration := flag.Duration("duration", 0, "overrides the duration")
	iterations := flag.Int("iterations", 0, "overrides the iterations per user")
	rampUp := flag.Duration("rampup", 0, "overrides the ramp-up duration")
	report := flag.String("report", "", "writes the report as JSON into the given file")
	flag.Parse()

	if *file == "" {
		flag.Usage()
		return 2
	}

	f, err := load.LoadFile(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	cfg := f.Config()
	flag.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "url":
			cfg.URL = *url
		case "users":
			cfg.Users = *users
		case "duration":
			cfg.Duration = *duration
		case "iterations":
			cfg.Iterations = *iterations
		case "rampup":
			cfg.RampUp = *rampUp
		}
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Fprintf(os.Stderr, "running %d users against %s (stop with ctrl+c)\n", cfg.Users, cfg.URL)
	start := time.Now()
	rep, err := load.Run(ctx, cfg, f.ScenarioList()...)
	if rep == nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	fmt.Fprintf(os.Stderr, "completed after %v\n\n", time.Since(start).Round(time.Millisecond))
	_ = rep.WriteText(os.Stdout)

	if *report != "" {
		buf, jerr := json.MarshalIndent(rep, "", "  ")
		if jerr == nil {
			jerr = os.WriteFile(*report, buf, 0644)
		}

		if jerr != nil {
			fmt.Fprintln(os.Stderr, jerr)
			return 2
		}
	}

	switch {
	case errors.Is(err, load.ErrThresholds):
		return 1
	case err != nil:
		fmt.Fprintln(os.Stderr, err)
		return 2
	default:
		return 0
	}
}
