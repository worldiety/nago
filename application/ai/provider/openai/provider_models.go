// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package openai

import (
	"iter"
	"slices"
	"strings"

	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
)

// nonChatModelMarkers identify models which a /models endpoint lists but which cannot be used with the Chat
// Completions API, like embedding, speech or image models. The list is a heuristic over the well-known naming
// schemes of OpenAI and of the common local model libraries.
var nonChatModelMarkers = []string{
	"embed",
	"whisper",
	"tts",
	"transcribe",
	"dall-e",
	"gpt-image",
	"moderation",
	"realtime",
	"sora",
	"davinci-002",
	"babbage-002",
}

func isChatModel(id string) bool {
	id = strings.ToLower(id)
	for _, marker := range nonChatModelMarkers {
		if strings.Contains(id, marker) {
			return false
		}
	}
	return true
}

// listModels is the implementation of completion.Completions.Models. The models are sorted by id, and models
// which are recognizably not chat models are left out.
func (p *openaiProvider) listModels(subject auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) {
		models, err := p.client().ListModels()
		if err != nil {
			yield(model.Model{}, err)
			return
		}

		slices.SortFunc(models, func(a, b apiModel) int {
			return strings.Compare(a.ID, b.ID)
		})

		for _, m := range models {
			if m.ID == "" || !isChatModel(m.ID) {
				continue
			}

			var desc string
			if m.OwnedBy != "" {
				desc = "owned by " + m.OwnedBy
			}

			if !yield(model.Model{ID: model.ID(m.ID), Name: m.ID, Description: desc}, nil) {
				return
			}
		}
	}
}
