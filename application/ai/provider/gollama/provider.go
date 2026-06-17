// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package gollama implements a local AI provider backed by llama.cpp through the gollama.cpp binding.
//
// It exposes the stateless [completion.Completions] capability (analogous to the anthropic provider) so the
// same agentic loop works against locally executed GGUF models. The provider assumes that llama.cpp itself is
// available (the gollama binding loads/downloads the native library on demand) and that GGUF model files
// either live in a configured search folder or can be downloaded from HuggingFace into a storage folder.
//
// Because different GGUF models require different prompt formats, tool-call encodings and detokenization
// schemes, those quirks are isolated behind the internal [adapter] and [pieceDecoder] abstractions, selected
// per model from its GGUF metadata (see gguf.go, adapter.go).
package gollama

import (
	"iter"
	"os"
	"path/filepath"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/tool"
	"go.wdy.de/nago/auth"
)

var _ provider.Provider = (*gollamaProvider)(nil)

type gollamaProvider struct {
	id          provider.ID
	cfg         Settings
	eng         *engine
	models      *gollamaModels
	completions *gollamaCompletions
}

// NewProvider creates a local llama.cpp provider. Only Models, Tools and Completions are supported; the
// stateful capabilities (Libraries, Agents, Conversations, Files) are unavailable because llama.cpp is
// stateless from the caller's perspective.
func NewProvider(id provider.ID, cfg Settings) provider.Provider {
	p := &gollamaProvider{
		id:  id,
		cfg: cfg,
		eng: newEngine(cfg),
	}

	p.models = &gollamaModels{parent: p}
	p.completions = &gollamaCompletions{parent: p}

	return p
}

func (p *gollamaProvider) Identity() provider.ID {
	return p.id
}

func (p *gollamaProvider) Name() string {
	return p.cfg.Name
}

func (p *gollamaProvider) Description() string {
	return p.cfg.Description
}

func (p *gollamaProvider) Models() provider.Models {
	return p.models
}

func (p *gollamaProvider) Tools() provider.Tools {
	return gollamaTools{}
}

func (p *gollamaProvider) Completions() option.Opt[completion.Completions] {
	return option.Some[completion.Completions](p.completions)
}

// ----- intentionally unsupported stateful capabilities -----

func (p *gollamaProvider) Libraries() option.Opt[provider.Libraries] {
	return option.None[provider.Libraries]()
}

func (p *gollamaProvider) Agents() option.Opt[provider.Agents] {
	return option.None[provider.Agents]()
}

func (p *gollamaProvider) Conversations() option.Opt[provider.Conversations] {
	return option.None[provider.Conversations]()
}

func (p *gollamaProvider) Files() option.Opt[provider.Files] {
	return option.None[provider.Files]()
}

// gollamaTools reports no parameterless built-in tools. Function tools are supplied per request via
// completion.Options.Tools instead.
type gollamaTools struct{}

func (gollamaTools) All(subject auth.Subject) iter.Seq2[tool.Tool, error] {
	return func(yield func(tool.Tool, error) bool) {}
}

// searchDir resolves the effective folder scanned for existing models.
func (s Settings) searchDir() string {
	if s.SearchDir != "" {
		return s.SearchDir
	}
	return defaultModelDir()
}

// storageDir resolves the effective folder that missing models are downloaded into.
func (s Settings) storageDir() string {
	if s.StorageDir != "" {
		return s.StorageDir
	}
	return defaultModelDir()
}

// defaultModelDir is the built-in fallback location for GGUF models. Models are re-downloadable blobs, so the
// user cache directory is an appropriate default. It falls back to the OS temp dir if no cache dir is known.
func defaultModelDir() string {
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "nago", "ai", "gollama", "models")
}
