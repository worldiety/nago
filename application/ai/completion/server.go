// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

// ServerToolKind names a tool which the provider runs itself, see [ServerTool].
type ServerToolKind string

const (
	// WebSearch searches the web and returns the hits with their titles and urls.
	WebSearch ServerToolKind = "web_search"
	// WebFetch reads a web page or a PDF, whose url is already part of the conversation: written by the user,
	// returned by a tool or found by a former search or fetch.
	WebFetch ServerToolKind = "web_fetch"
)

// ServerTool is a tool which the provider runs itself, within the same request: the model calls it, the
// provider executes it and the model continues with the result, all before the answer arrives. Nothing is
// executed by nago and no [ToolResult] is sent back. The answer carries the call and its result as
// [ServerBlock]s instead, which must stay in the history unchanged.
//
// The Anthropic provider passes them to the Messages API, which offers them natively, and the Nago AI Service
// also runs them for its models of other providers. Providers without server-side tools ignore them.
type ServerTool struct {
	Kind ServerToolKind `json:"kind"`
	// MaxUses limits how often the model may call the tool within one turn, i.e. until it answers. Zero means
	// the default of the provider.
	MaxUses int `json:"maxUses,omitzero"`
	// AllowedDomains restricts the tool to these domains, e.g. example.com, which includes its subdomains.
	// Optional; it must not be combined with BlockedDomains.
	AllowedDomains []string `json:"allowedDomains,omitzero"`
	// BlockedDomains excludes these domains. Optional.
	BlockedDomains []string `json:"blockedDomains,omitzero"`
	// MaxContentTokens truncates a fetched page to about this many tokens. Only for [WebFetch]; zero means
	// the default of the provider.
	MaxContentTokens int `json:"maxContentTokens,omitzero"`
}

// Defaults of [WebResearchTools].
const (
	DefaultWebSearchMaxUses   = 5
	DefaultWebFetchMaxUses    = 10
	DefaultWebFetchMaxContent = 25000
)

// WebResearchTools returns the server tools of a web research: a search and the fetch of the pages it found.
// Per turn, the model may search [DefaultWebSearchMaxUses] times and read [DefaultWebFetchMaxUses] pages of up
// to about [DefaultWebFetchMaxContent] tokens each.
func WebResearchTools() []ServerTool {
	return []ServerTool{
		{Kind: WebSearch, MaxUses: DefaultWebSearchMaxUses},
		{Kind: WebFetch, MaxUses: DefaultWebFetchMaxUses, MaxContentTokens: DefaultWebFetchMaxContent},
	}
}

// ServerBlock is a content block of a [ServerTool] in an assistant turn: its call or its result, as the provider
// returned it. Raw is the block in the format of the Anthropic Messages API, which the Nago AI Service also
// uses for the server tools it runs itself. It must be sent back exactly as received and in its place, because
// the provider validates it, e.g. the encrypted content of a search result. A provider which cannot replay it
// gets its text instead, see [FlattenServerBlocks].
//
// A call without its result in the same turn is a deferred one: it happens when the model calls a server tool
// together with a client [Tool]. The result then opens the next answer of the model.
type ServerBlock struct {
	Raw json.RawMessage `json:"raw"`
}

func (ServerBlock) isContent() {}

// ServerCall is the call of a server tool, see [ServerBlock.Call].
type ServerCall struct {
	ID   string
	Kind ServerToolKind
	// Query is what the model searches for, URL what it fetches.
	Query string
	URL   string
}

// WebSource is a page a server tool found or fetched.
type WebSource struct {
	URL   string
	Title string
	// Age is how old the page is, as the search reports it, e.g. "April 30, 2025". Optional.
	Age string
}

// Sources returns the pages the text cites, in the order of their first citation. A citation without an address,
// like one of a document, is left out.
func (t Text) Sources() []WebSource {
	var cites []struct {
		URL   string `json:"url"`
		Title string `json:"title"`
	}
	if len(t.Citations) == 0 || json.Unmarshal(t.Citations, &cites) != nil {
		return nil
	}

	var res []WebSource
	for _, c := range cites {
		if c.URL == "" || slices.ContainsFunc(res, func(s WebSource) bool { return s.URL == c.URL }) {
			continue
		}
		res = append(res, WebSource{URL: c.URL, Title: c.Title})
	}

	return res
}

// ServerResult is the result of a server tool, see [ServerBlock.Result].
type ServerResult struct {
	// CallID is the id of the [ServerCall].
	CallID string
	Kind   ServerToolKind
	// ErrorCode is set, if the tool failed, e.g. url_not_accessible or max_uses_exceeded.
	ErrorCode string
	// Sources are the hits of a search or the fetched page.
	Sources []WebSource
	// Text is the fetched text. A fetched PDF has none.
	Text string
}

// Type returns the block type, e.g. server_tool_use or web_search_tool_result.
func (b ServerBlock) Type() string {
	var head struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal(b.Raw, &head)
	return head.Type
}

// Call returns the call, if the block is one.
func (b ServerBlock) Call() (ServerCall, bool) {
	var w struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Name  string `json:"name"`
		Input struct {
			Query string `json:"query"`
			URL   string `json:"url"`
		} `json:"input"`
	}
	if json.Unmarshal(b.Raw, &w) != nil || w.Type != "server_tool_use" {
		return ServerCall{}, false
	}

	return ServerCall{ID: w.ID, Kind: ServerToolKind(w.Name), Query: w.Input.Query, URL: w.Input.URL}, true
}

// Result returns the result of a web search or a web fetch, if the block is one.
func (b ServerBlock) Result() (ServerResult, bool) {
	var w struct {
		Type      string          `json:"type"`
		ToolUseID string          `json:"tool_use_id"`
		Content   json.RawMessage `json:"content"`
	}
	if json.Unmarshal(b.Raw, &w) != nil {
		return ServerResult{}, false
	}

	res := ServerResult{CallID: w.ToolUseID}
	switch w.Type {
	case "web_search_tool_result":
		res.Kind = WebSearch
	case "web_fetch_tool_result":
		res.Kind = WebFetch
	default:
		return ServerResult{}, false
	}

	// an error is a single object, a search a list of hits and a fetch a single document
	var obj struct {
		Type      string `json:"type"`
		ErrorCode string `json:"error_code"`
		URL       string `json:"url"`
		Content   struct {
			Title  string `json:"title"`
			Source struct {
				Type string `json:"type"`
				Data string `json:"data"`
			} `json:"source"`
		} `json:"content"`
	}
	if json.Unmarshal(w.Content, &obj) == nil {
		switch obj.Type {
		case "web_search_tool_result_error", "web_fetch_tool_result_error":
			res.ErrorCode = obj.ErrorCode
			if res.ErrorCode == "" {
				res.ErrorCode = "unavailable"
			}
		case "web_fetch_result":
			res.Sources = []WebSource{{URL: obj.URL, Title: obj.Content.Title}}
			if obj.Content.Source.Type == "text" {
				res.Text = obj.Content.Source.Data
			}
		}
		return res, true
	}

	var hits []struct {
		Type    string `json:"type"`
		URL     string `json:"url"`
		Title   string `json:"title"`
		PageAge string `json:"page_age"`
	}
	if json.Unmarshal(w.Content, &hits) == nil {
		for _, h := range hits {
			if h.URL != "" {
				res.Sources = append(res.Sources, WebSource{URL: h.URL, Title: h.Title, Age: h.PageAge})
			}
		}
	}

	return res, true
}

// Summary describes the block as text, for a transcript or a provider which cannot replay it. The text of a
// fetched page is cut to maxRunes; zero leaves it out.
func (b ServerBlock) Summary(maxRunes int) string {
	if call, ok := b.Call(); ok {
		switch {
		case call.Query != "":
			return fmt.Sprintf("[%s %q]", call.Kind, call.Query)
		case call.URL != "":
			return fmt.Sprintf("[%s %s]", call.Kind, call.URL)
		default:
			return fmt.Sprintf("[%s]", call.Kind)
		}
	}

	res, ok := b.Result()
	if !ok {
		return "[" + b.Type() + "]"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "[%s result", res.Kind)
	if res.ErrorCode != "" {
		fmt.Fprintf(&sb, " error %s]", res.ErrorCode)
		return sb.String()
	}

	sb.WriteString("]")
	for _, s := range res.Sources {
		sb.WriteString("\n- ")
		if s.Title != "" {
			sb.WriteString(s.Title)
			sb.WriteString(": ")
		}
		sb.WriteString(s.URL)
	}

	if text := strings.TrimSpace(res.Text); text != "" && maxRunes > 0 {
		if utf8.RuneCountInString(text) > maxRunes {
			text = string([]rune(text)[:maxRunes]) + " …"
		}
		sb.WriteString("\n")
		sb.WriteString(text)
	}

	return sb.String()
}

// flattenedRunes is how much of a fetched page [FlattenServerBlocks] keeps.
const flattenedRunes = 8000

// FlattenServerBlocks returns the history with every [ServerBlock] replaced by its [ServerBlock.Summary] and
// without the citations of the texts. It is for providers which cannot replay the blocks, e.g. after the
// conversation moved to another provider, and keeps what was found readable for the model. With kinds, only the
// blocks of these tools are replaced, and the citations only with the search. The input is not modified.
func FlattenServerBlocks(history []Message, kinds ...ServerToolKind) []Message {
	if !HasServerBlocks(history) {
		return history
	}

	flatten := func(k ServerToolKind) bool { return len(kinds) == 0 || slices.Contains(kinds, k) }
	out := make([]Message, len(history))
	for i, m := range history {
		content := make([]Content, 0, len(m.Content))
		for _, c := range m.Content {
			switch v := c.(type) {
			case ServerBlock:
				if flatten(v.Kind()) {
					c = Text{Text: v.Summary(flattenedRunes)}
				}
			case Text:
				if flatten(WebSearch) {
					c = Text{Text: v.Text}
				}
			}
			content = append(content, c)
		}
		out[i] = Message{Role: m.Role, Content: content}
	}

	return out
}

// Kind returns the tool of the call or the result, or an empty kind for another block.
func (b ServerBlock) Kind() ServerToolKind {
	if call, ok := b.Call(); ok {
		return call.Kind
	}

	if res, ok := b.Result(); ok {
		return res.Kind
	}

	return ""
}

// HasServerBlocks reports whether the history contains a [ServerBlock] or a text with citations.
func HasServerBlocks(history []Message) bool {
	for _, m := range history {
		for _, c := range m.Content {
			switch v := c.(type) {
			case ServerBlock:
				return true
			case Text:
				if len(v.Citations) > 0 {
					return true
				}
			}
		}
	}

	return false
}

// pendingServerCall reports whether the assistant turn calls a server tool without its result, see
// [ServerBlock]. The provider then delivers the result with its next answer and refuses a follow-up user turn
// with anything but tool results.
func pendingServerCall(msg Message) bool {
	calls := map[string]bool{}
	for _, c := range msg.Content {
		b, ok := c.(ServerBlock)
		if !ok {
			continue
		}

		if call, ok := b.Call(); ok {
			calls[call.ID] = true
		} else if res, ok := b.Result(); ok {
			delete(calls, res.CallID)
		}
	}

	return len(calls) > 0
}

// withoutPendingServerCalls removes the server calls of msg which have no result in it, see
// [pendingServerCall]. A truncated turn must not keep them, because nobody delivers their results.
func withoutPendingServerCalls(msg Message) Message {
	answered := map[string]bool{}
	for _, c := range msg.Content {
		if b, ok := c.(ServerBlock); ok {
			if res, ok := b.Result(); ok {
				answered[res.CallID] = true
			}
		}
	}

	out := make([]Content, 0, len(msg.Content))
	for _, c := range msg.Content {
		if b, ok := c.(ServerBlock); ok {
			if call, ok := b.Call(); ok && !answered[call.ID] {
				continue
			}
		}
		out = append(out, c)
	}

	return Message{Role: msg.Role, Content: out}
}
