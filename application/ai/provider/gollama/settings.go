// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"github.com/worldiety/enum"
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/secret"
	"golang.org/x/text/language"
)

var _ = enum.Variant[secret.Credentials, Settings](
	enum.Rename[Settings]("nago.ai.gollama.settings"),
)

var (
	StrSettingsTitle           = i18n.MustString("nago.ai.gollama.settings_title", i18n.Values{language.English: "My local llama.cpp models", language.German: "Meine lokalen llama.cpp Modelle"})
	StrSettingsName            = i18n.MustString("nago.ai.gollama.settings_name", i18n.Values{language.English: "Local (llama.cpp)", language.German: "Lokal (llama.cpp)"})
	StrSettingsDescription     = i18n.MustString("nago.ai.gollama.settings_desc", i18n.Values{language.English: "Run GGUF models locally via llama.cpp (gollama).", language.German: "GGUF-Modelle lokal über llama.cpp (gollama) ausführen."})
	StrSettingsSearchDir       = i18n.MustString("nago.ai.gollama.settings_search_dir", i18n.Values{language.English: "Model search folder", language.German: "Modell-Suchordner"})
	StrSettingsSearchDirDesc   = i18n.MustString("nago.ai.gollama.settings_search_dir_desc", i18n.Values{language.English: "Folder that is scanned for already present *.gguf model files. Leave empty for the built-in default.", language.German: "Ordner, der nach bereits vorhandenen *.gguf Modelldateien durchsucht wird. Leer lassen für den Standardwert."})
	StrSettingsStorageDir      = i18n.MustString("nago.ai.gollama.settings_storage_dir", i18n.Values{language.English: "Model download folder", language.German: "Modell-Downloadordner"})
	StrSettingsStorageDirDesc  = i18n.MustString("nago.ai.gollama.settings_storage_dir_desc", i18n.Values{language.English: "Target folder where missing models are downloaded to (e.g. from HuggingFace). Leave empty for the built-in default.", language.German: "Zielordner, in den fehlende Modelle heruntergeladen werden (z.B. von HuggingFace). Leer lassen für den Standardwert."})
	StrSettingsHFToken         = i18n.MustString("nago.ai.gollama.settings_hf_token", i18n.Values{language.English: "HuggingFace token", language.German: "HuggingFace Token"})
	StrSettingsHFTokenDesc     = i18n.MustString("nago.ai.gollama.settings_hf_token_desc", i18n.Values{language.English: "Optional access token used to download gated or private models from HuggingFace.", language.German: "Optionaler Zugriffstoken zum Herunterladen geschützter oder privater Modelle von HuggingFace."})
	StrSettingsContextSize     = i18n.MustString("nago.ai.gollama.settings_ctx", i18n.Values{language.English: "Default context size", language.German: "Standard-Kontextgröße"})
	StrSettingsContextSizeDesc = i18n.MustString("nago.ai.gollama.settings_ctx_desc", i18n.Values{language.English: "Context window in tokens, always capped at the model's maximum. 0 = use the built-in default (4096).", language.German: "Kontextfenster in Tokens, stets begrenzt auf das Maximum des Modells. 0 = Standardwert verwenden (4096)."})
	StrSettingsThreads         = i18n.MustString("nago.ai.gollama.settings_threads", i18n.Values{language.English: "CPU threads", language.German: "CPU-Threads"})
	StrSettingsThreadsDesc     = i18n.MustString("nago.ai.gollama.settings_threads_desc", i18n.Values{language.English: "Number of CPU threads used for generation. 0 = auto (number of cores).", language.German: "Anzahl der für die Generierung genutzten CPU-Threads. 0 = automatisch (Anzahl der Kerne)."})
	StrSettingsGpuLayers       = i18n.MustString("nago.ai.gollama.settings_gpu_layers", i18n.Values{language.English: "GPU layers", language.German: "GPU-Layer"})
	StrSettingsGpuLayersDesc   = i18n.MustString("nago.ai.gollama.settings_gpu_layers_desc", i18n.Values{language.English: "Number of model layers offloaded to the GPU. 0 = CPU only, a large value (e.g. 999) offloads everything.", language.German: "Anzahl der auf die GPU ausgelagerten Modell-Layer. 0 = nur CPU, ein großer Wert (z.B. 999) lagert alles aus."})
)

// Settings configures the local llama.cpp (gollama) provider. It is a [secret.Credentials] so administrators
// can manage it like any other AI provider. Unlike the cloud providers there is no mandatory API token; the
// only secret field is the optional HuggingFace token used for gated model downloads.
//
// The set of available models is NOT configured here. It is a curated, hardcoded catalog in catalog.go,
// because each entry needs to be reviewed and tested against the target machines.
type Settings struct {
	Name        string `value:"nago.ai.gollama.settings_title"`
	Description string `label:"nago.common.label.description" lines:"3"`

	// SearchDir is the folder scanned for already present *.gguf files. Empty means the provider falls back
	// to its built-in default (see [Settings.searchDir]).
	SearchDir string `label:"nago.ai.gollama.settings_search_dir" supportingText:"nago.ai.gollama.settings_search_dir_desc" json:"searchDir"`

	// StorageDir is the folder that missing models are downloaded into. Empty means the built-in default.
	StorageDir string `label:"nago.ai.gollama.settings_storage_dir" supportingText:"nago.ai.gollama.settings_storage_dir_desc" json:"storageDir"`

	// HFToken is the optional HuggingFace access token for gated/private repositories.
	HFToken string `label:"nago.ai.gollama.settings_hf_token" supportingText:"nago.ai.gollama.settings_hf_token_desc" style:"secret" json:"hfToken"`

	// CtxSize is the default context window in tokens. 0 derives it from the model. A catalog entry may
	// override it.
	CtxSize int `label:"nago.ai.gollama.settings_ctx" supportingText:"nago.ai.gollama.settings_ctx_desc" json:"ctxSize"`

	// Threads is the number of CPU threads. 0 selects the number of logical cores.
	Threads int `label:"nago.ai.gollama.settings_threads" supportingText:"nago.ai.gollama.settings_threads_desc" json:"threads"`

	// GpuLayers is the number of layers offloaded to the GPU. 0 keeps everything on the CPU.
	GpuLayers int `label:"nago.ai.gollama.settings_gpu_layers" supportingText:"nago.ai.gollama.settings_gpu_layers_desc" json:"gpuLayers"`

	Debug bool `json:"debug"`

	_ struct{} `credentialName:"nago.ai.gollama.settings_name" credentialDescription:"nago.ai.gollama.settings_desc"`
}

func (Settings) Credentials() bool {
	return true
}

func (s Settings) GetName() string {
	return s.Name
}

func (s Settings) IsZero() bool {
	return s == Settings{}
}
