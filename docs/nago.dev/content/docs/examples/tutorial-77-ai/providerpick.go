// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"

	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

// firstCompletionProvider returns the first configured provider that exposes stateless completions, together
// with its Completions capability. It is a small convenience used by the uicompletion-based demo pages, which
// (unlike the low-level examples) do not offer a provider dropdown but simply take the first usable provider.
// requireFiles additionally demands a Files capability (needed for file upload / drive tools).
func firstCompletionProvider(subject auth.Subject, uc ai.UseCases, requireFiles bool) (provider.Provider, completion.Completions, error) {
	for p, err := range uc.FindAllProvider(subject) {
		if err != nil {
			return nil, nil, err
		}

		c := p.Completions()
		if c.IsNone() {
			continue
		}
		if requireFiles && p.Files().IsNone() {
			continue
		}
		return p, c.Unwrap(), nil
	}

	if requireFiles {
		return nil, nil, fmt.Errorf("kein Provider mit Completions UND Files gefunden – bitte ein Anthropic-Secret konfigurieren")
	}
	return nil, nil, fmt.Errorf("kein Provider mit stateless Completions gefunden – bitte ein Secret konfigurieren")
}
