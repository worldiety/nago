// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cache

import (
	"iter"
	"sync"

	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/tool"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xslices"
)

var _ provider.Tools = (*cacheTools)(nil)

type cacheTools struct {
	parent *Provider
	mu     sync.Mutex
	loaded bool
	tools  []tool.Tool
}

func (c *cacheTools) All(subject auth.Subject) iter.Seq2[tool.Tool, error] {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loaded {
		return xslices.ValuesWithError(c.tools, nil)
	}

	return func(yield func(tool.Tool, error) bool) {
		c.mu.Lock()
		defer c.mu.Unlock()

		tools, err := xslices.Collect2(c.parent.prov.Tools().All(subject))
		if err != nil {
			yield(tool.Tool{}, err)
			return
		}

		c.tools = tools
		c.loaded = true

		for _, t := range tools {
			if !yield(t, nil) {
				return
			}
		}
	}
}
