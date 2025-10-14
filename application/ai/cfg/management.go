// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/agent"
	uidrive "go.wdy.de/nago/application/drive/ui"
	"go.wdy.de/nago/presentation/core"
)

type Management struct {
	AgentUseCases agent.UseCases
	Pages         uidrive.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	panic("todo")
}
