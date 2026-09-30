// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package openai

import (
	"github.com/worldiety/enum"
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/secret"
	"golang.org/x/text/language"
)

// The rename is part of the persisted secret format and must never change, otherwise stored OpenAI secrets can
// no longer be decoded.
var _ = enum.Variant[secret.Credentials, Settings](
	enum.Rename[Settings]("nago.ai.openai.settings"),
)

// Register wires this provider's factory into the global provider registry, so it becomes available only when
// the host application side-imports this package.
var _ = registerProvider()

func registerProvider() any {
	provider.Register[Settings](NewProvider)
	return nil
}

var (
	StrOpenAISettingsTitle         = i18n.MustString("nago.ai.openai.settings_title", i18n.Values{language.English: "My OpenAI Token", language.German: "Mein OpenAI Token"})
	StrOpenAISettingsName          = i18n.MustString("nago.ai.openai.settings_name", i18n.Values{language.English: "OpenAI (compatible)", language.German: "OpenAI (kompatibel)"})
	StrOpenAISettingsDescription   = i18n.MustString("nago.ai.openai.settings_desc", i18n.Values{language.English: "Connect to OpenAI or any OpenAI-compatible server (Ollama, vLLM, LM Studio, OpenRouter, llama.cpp server, ...) via the Chat Completions API", language.German: "Anbindung von OpenAI oder eines OpenAI-kompatiblen Servers (Ollama, vLLM, LM Studio, OpenRouter, llama.cpp Server, ...) über die Chat Completions API"})
	StrOpenAISettingsToken         = i18n.MustString("nago.ai.openai.settings_token", i18n.Values{language.English: "API Token", language.German: "API Token"})
	StrOpenAISettingsTokenDesc     = i18n.MustString("nago.ai.openai.settings_token_desc", i18n.Values{language.English: "Sent as Bearer token. Local servers like Ollama usually do not need one, so it may be left empty.", language.German: "Wird als Bearer-Token gesendet. Lokale Server wie Ollama benötigen meist keinen, dann kann das Feld leer bleiben."})
	StrOpenAISettingsBaseURL       = i18n.MustString("nago.ai.openai.settings_base_url", i18n.Values{language.English: "Base URL", language.German: "Basis-URL"})
	StrOpenAISettingsBaseURLDesc   = i18n.MustString("nago.ai.openai.settings_base_url_desc", i18n.Values{language.English: "Endpoint of the API including the version path, e.g. http://localhost:11434/v1 for Ollama. Leave empty for https://api.openai.com/v1.", language.German: "Endpunkt der API inklusive Versionspfad, z.B. http://localhost:11434/v1 für Ollama. Leer lassen für https://api.openai.com/v1."})
	StrOpenAISettingsMaxTokens     = i18n.MustString("nago.ai.openai.settings_max_tokens", i18n.Values{language.English: "Default max. output tokens", language.German: "Standard max. Ausgabe-Tokens"})
	StrOpenAISettingsMaxTokensDesc = i18n.MustString("nago.ai.openai.settings_max_tokens_desc", i18n.Values{language.English: "Used when a request does not set a limit itself. 0 = let the server decide.", language.German: "Wird verwendet, wenn eine Anfrage selbst kein Limit setzt. 0 = der Server entscheidet."})
	StrOpenAISettingsRPS           = i18n.MustString("nago.ai.openai.settings_rps", i18n.Values{language.English: "Requests per Second", language.German: "Anfragen pro Sekunde"})
	StrOpenAISettingsRPSDesc       = i18n.MustString("nago.ai.openai.settings_rps_desc", i18n.Values{language.English: "Limit the rate of requests against the API", language.German: "Anfragebegrenzung pro Sekunde an die API."})
)

// Settings configures the OpenAI provider. It targets the stateless Chat Completions API and therefore works
// with OpenAI itself as well as with OpenAI-compatible servers, by pointing BaseURL at them.
//
// The Name and Token fields intentionally keep their historic (untagged) JSON names, so secrets stored by earlier
// versions of this package decode unchanged.
type Settings struct {
	Name        string `value:"nago.ai.openai.settings_title"`
	Description string `label:"nago.common.label.description" lines:"3"`
	// Token is the API key sent as Bearer token. Optional, because local servers usually do not require one.
	Token string `label:"nago.ai.openai.settings_token" supportingText:"nago.ai.openai.settings_token_desc" style:"secret"`
	// BaseURL is the API endpoint including the version path. Empty means [DefaultBaseURL].
	BaseURL string `label:"nago.ai.openai.settings_base_url" supportingText:"nago.ai.openai.settings_base_url_desc" json:"baseUrl"`
	// MaxTokens is the output token limit used when [completion.Options.MaxTokens] is not set. Zero omits the
	// limit, so the server default applies.
	MaxTokens int      `label:"nago.ai.openai.settings_max_tokens" supportingText:"nago.ai.openai.settings_max_tokens_desc" json:"maxTokens"`
	RPS       int      `label:"nago.ai.openai.settings_rps" supportingText:"nago.ai.openai.settings_rps_desc" json:"rps"`
	Debug     bool     `json:"debug"`
	_         struct{} `credentialName:"nago.ai.openai.settings_name" credentialDescription:"nago.ai.openai.settings_desc" credentialLogo:"https://openai.com/favicon.svg"`
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
