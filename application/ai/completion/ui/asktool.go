// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import "go.wdy.de/nago/application/ai/completion"

// askUserToolName is the name of the built-in clarification tool (see [completion.NewAskUserTool]). The
// renderer uses it to show the question and the user's answer as regular chat bubbles.
const askUserToolName = completion.AskUserToolName
