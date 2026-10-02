// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Command screenshots renders the screenshots of nago.dev reproducibly. It builds each example listed in
// shots.yaml, starts it with an empty data directory, drives it with a headless Chrome and writes the
// cropped images into the Hugo site. See README.md.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	manifest := flag.String("f", "shots.yaml", "the manifest which declares all screenshots")
	only := flag.String("only", "", "comma separated list of examples to render, e.g. tutorial-11-buttons")
	headful := flag.Bool("headful", false, "show the browser window, useful to debug steps")
	flag.Parse()

	m, err := loadManifest(*manifest)
	if err != nil {
		log.Fatal(err)
	}

	root, err := filepath.Abs("../..")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(root, "example", "cmd")); err != nil {
		log.Fatalf("must be run from docs/screenshots: %v", err)
	}

	filter := map[string]bool{}
	for _, s := range strings.Split(*only, ",") {
		if s = strings.TrimSpace(s); s != "" {
			filter[s] = true
		}
	}

	r := &runner{
		root:     root,
		site:     filepath.Join(root, "docs", "nago.dev"),
		manifest: m,
		headful:  *headful,
	}

	failed := 0
	for _, group := range m.groupByExample() {
		if len(filter) > 0 && !filter[group.example] {
			continue
		}

		if err := r.renderExample(context.Background(), group.example, group.shots); err != nil {
			log.Printf("%s: %v", group.example, err)
			failed++
		}
	}

	if failed > 0 {
		fmt.Fprintf(os.Stderr, "%d example(s) failed\n", failed)
		os.Exit(1)
	}
}
