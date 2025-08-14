// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewDeclare(declarations *concurrent.RWMap[ID, *workflow]) Declare {
	return func(subject user.Subject, opts DeclareOptions) (ID, error) {
		if opts.ID == "" {
			opts.ID = data.RandIdent[ID]()
		}

		if strings.Contains(string(opts.ID), "/") {
			return "", fmt.Errorf("id must not contain '/'")
		}

		if _, ok := declarations.Get(opts.ID); ok {
			return "", fmt.Errorf("already declared")
		}

		if len(opts.Actions) == 0 {
			return "", fmt.Errorf("no actions")
		}

		wf := newWorkflow(opts)

		for _, action := range opts.Actions {
			cfg := newConfiguration(opts.ID)
			if err := action.Configure(cfg); err != nil {
				return "", fmt.Errorf("cannot configure action `%v`: %w", action, err)
			}
			wf.nodes = append(wf.nodes, newNode(cfg, action))
		}

		wf.init()

		declarations.Put(opts.ID, wf)
		return opts.ID, nil
	}
}
