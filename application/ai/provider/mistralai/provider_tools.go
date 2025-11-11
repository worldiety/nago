// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"iter"

	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/tool"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xslices"
)

var _ provider.Tools = (*mistralTools)(nil)

type mistralTools struct {
}

func (m mistralTools) All(subject auth.Subject) iter.Seq2[tool.Tool, error] {
	models := []tool.Tool{
		{
			ID:          "web_search",
			Name:        "Web Search",
			Description: "A simple web search tool that enables access to a search engine.",
		},

		{
			ID:          "web_search_premium",
			Name:        "Web Search Premium",
			Description: "A more complex web search tool that enables access to both a search engine and to news articles via integrated news provider verification.",
		},

		{
			ID:          "code_interpreter",
			Name:        "Code Interpreter",
			Description: "To use the code interpreter, you can create an agent with the code interpreter tool, once done you can start a conversation with the agent and it will run code on demand, leveraging the outputs to answer your questions.",
		},
		{
			ID:          "image_generation",
			Name:        "Image Generation",
			Description: "Image Generation is a built-in tool tool that enables agents to generate images of all kinds and forms.",
		},

		// note that the library has a more complex model, because we need the entire library management
	}

	return xslices.ValuesWithError(models, nil)
}
