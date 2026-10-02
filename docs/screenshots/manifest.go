// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Defaults Shot   `yaml:"defaults"`
	Shots    []Shot `yaml:"shots"`
}

// Shot declares a single image. Zero values are taken from the manifest defaults.
type Shot struct {
	// Example is the directory name below example/cmd or a package path relative to the repository root,
	// like example/gallery/basic.
	Example string `yaml:"example"`
	// Out is the target file, relative to docs/nago.dev. The extension decides the format (.webp or .png).
	Out string `yaml:"out"`
	// Path is the url path to open, defaults to /.
	Path string `yaml:"path"`
	// Width and Height of the browser viewport in CSS pixels.
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
	// Scale is the device pixel ratio.
	Scale float64 `yaml:"scale"`
	// Dark emulates prefers-color-scheme: dark.
	Dark *bool `yaml:"dark"`
	// Crop is one of auto (bounding box of the visible content), viewport, page or a css selector.
	Crop string `yaml:"crop"`
	// Padding around an auto or selector crop in CSS pixels.
	Padding *int `yaml:"padding"`
	// Quality for lossy formats, 1-100.
	Quality int `yaml:"quality"`
	// Admin enables the bootstrap admin admin@localhost before the example starts. Steps can refer to its
	// random password as {{adminPassword}}. If any shot of an example sets it, it applies to all of them.
	Admin bool `yaml:"admin"`
	// Steps are executed after the page has settled and before the capture.
	Steps []Step `yaml:"steps"`
}

// Step is a single interaction. Exactly one field must be set.
type Step struct {
	// Click clicks the first element matching the css selector.
	Click string `yaml:"click"`
	// ClickText clicks the innermost element whose text equals the given text.
	ClickText string `yaml:"clickText"`
	// Type focuses the selector and types the text.
	Type *TypeStep `yaml:"type"`
	// Hover moves the mouse over the first element matching the css selector.
	Hover string `yaml:"hover"`
	// Wait sleeps, e.g. 500ms.
	Wait string `yaml:"wait"`
	// JS evaluates arbitrary JavaScript in the page.
	JS string `yaml:"js"`
	// Goto navigates to another url path of the same example.
	Goto string `yaml:"goto"`
	// Login signs in as the bootstrap admin, requires admin: true. The session survives for all following
	// shots of the same example.
	Login bool `yaml:"login"`
}

type TypeStep struct {
	Selector string `yaml:"selector"`
	Text     string `yaml:"text"`
}

func loadManifest(fname string) (Manifest, error) {
	buf, err := os.ReadFile(fname)
	if err != nil {
		return Manifest{}, err
	}

	var m Manifest
	if err := yaml.Unmarshal(buf, &m); err != nil {
		return Manifest{}, fmt.Errorf("invalid manifest %s: %w", fname, err)
	}

	for i, s := range m.Shots {
		if s.Example == "" || s.Out == "" {
			return Manifest{}, fmt.Errorf("shot #%d: example and out are required", i)
		}

		for _, step := range s.Steps {
			if step.Wait != "" {
				if _, err := time.ParseDuration(step.Wait); err != nil {
					return Manifest{}, fmt.Errorf("shot %s: invalid wait: %w", s.Out, err)
				}
			}
		}

		m.Shots[i] = m.withDefaults(s)
	}

	return m, nil
}

func (m Manifest) withDefaults(s Shot) Shot {
	d := m.Defaults
	if s.Path == "" {
		s.Path = or(d.Path, "/")
	}
	if s.Width == 0 {
		s.Width = or(d.Width, 1200)
	}
	if s.Height == 0 {
		s.Height = or(d.Height, 800)
	}
	if s.Scale == 0 {
		s.Scale = or(d.Scale, 2)
	}
	if s.Dark == nil {
		s.Dark = d.Dark
	}
	if s.Crop == "" {
		s.Crop = or(d.Crop, "auto")
	}
	if s.Padding == nil {
		s.Padding = d.Padding
	}
	if s.Padding == nil {
		p := 24
		s.Padding = &p
	}
	if s.Quality == 0 {
		s.Quality = or(d.Quality, 90)
	}

	return s
}

type exampleGroup struct {
	example string
	shots   []Shot
}

// groupByExample keeps the manifest order, so that each example is only built and started once.
func (m Manifest) groupByExample() []exampleGroup {
	var res []exampleGroup
	idx := map[string]int{}
	for _, s := range m.Shots {
		i, ok := idx[s.Example]
		if !ok {
			i = len(res)
			idx[s.Example] = i
			res = append(res, exampleGroup{example: s.Example})
		}

		res[i].shots = append(res[i].shots, s)
	}

	return res
}

func or[T comparable](v, fallback T) T {
	var zero T
	if v == zero {
		return fallback
	}

	return v
}
