// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import "go.wdy.de/nago/application/ai/model"

// family identifies the chat/tool quirks of a model. It selects the [adapter] used to render prompts and
// parse the generated output. An empty family means "auto-detect from the GGUF metadata".
type family string

const (
	familyAuto    family = ""        // detect from GGUF metadata (general.architecture / chat_template)
	familyChatML  family = "chatml"  // Qwen, Hermes and other ChatML models, Hermes-style tool calls
	familyLlama3  family = "llama3"  // Llama 3.x header format
	familyMistral family = "mistral" // Mistral / Mixtral [INST] format
)

// catalogEntry is a single, curated model offered by the local provider. The list is hardcoded (see
// [catalog]) on purpose: every entry has to be reviewed and tested against the target hardware. A model is
// looked up on disk by File in the configured search/storage folders and, if absent, downloaded from
// HuggingFace via HFRepo/HFFile.
type catalogEntry struct {
	// ID is the stable model identifier exposed through [completion.Options.Model].
	ID model.ID
	// Name is the human readable label.
	Name string
	// Description is shown in the admin UI.
	Description string
	// File is the GGUF filename looked up on disk and used as the download target name.
	File string
	// HFRepo is the HuggingFace repository, e.g. "Qwen/Qwen3-1.7B-GGUF".
	HFRepo string
	// HFFile is the file within HFRepo. Empty means it equals File.
	HFFile string
	// Family overrides the auto-detected chat/tool family. Leave empty to detect from GGUF metadata.
	Family family
	// CtxSize overrides the default context window for this model. 0 uses the provider/model default.
	CtxSize int
}

// hfFile returns the file name within the HuggingFace repository.
func (e catalogEntry) hfFile() string {
	if e.HFFile != "" {
		return e.HFFile
	}
	return e.File
}

func (e catalogEntry) toModel() model.Model {
	return model.Model{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
	}
}

// catalog is the curated set of locally runnable models. These entries are intentionally hardcoded and meant
// to be reviewed/adjusted for the concrete deployment. All referenced files are single-file GGUFs (no shards)
// so they can be downloaded and memory-mapped directly.
var catalog = []catalogEntry{
	{
		ID:          "qwen3-1.7b",
		Name:        "Qwen3 1.7B",
		Description: "Small, fast Qwen3 model. ChatML prompt, Hermes-style tool calls.",
		File:        "Qwen3-1.7B-Q8_0.gguf",
		HFRepo:      "Qwen/Qwen3-1.7B-GGUF",
		Family:      familyChatML,
	},
	{
		ID:          "qwen2.5-3b-instruct",
		Name:        "Qwen2.5 3B Instruct",
		Description: "Qwen2.5 instruct model with solid tool calling. ChatML prompt, Hermes-style tool calls.",
		File:        "qwen2.5-3b-instruct-q4_k_m.gguf",
		HFRepo:      "Qwen/Qwen2.5-3B-Instruct-GGUF",
		Family:      familyChatML,
	},
	{
		ID:          "llama-3.2-3b-instruct",
		Name:        "Llama 3.2 3B Instruct",
		Description: "Meta Llama 3.2 instruct model. Llama 3 header prompt and JSON tool calls.",
		File:        "Llama-3.2-3B-Instruct-Q4_K_M.gguf",
		HFRepo:      "bartowski/Llama-3.2-3B-Instruct-GGUF",
		Family:      familyLlama3,
	},
	{
		ID:          "mistral-7b-instruct-v0.3",
		Name:        "Mistral 7B Instruct v0.3",
		Description: "Mistral 7B instruct model. [INST] prompt and [TOOL_CALLS] tool format.",
		File:        "Mistral-7B-Instruct-v0.3-Q4_K_M.gguf",
		HFRepo:      "bartowski/Mistral-7B-Instruct-v0.3-GGUF",
		Family:      familyMistral,
	},
}

// lookupCatalog returns the catalog entry for the given model id.
func lookupCatalog(id model.ID) (catalogEntry, bool) {
	for _, e := range catalog {
		if e.ID == id {
			return e, true
		}
	}
	return catalogEntry{}, false
}
