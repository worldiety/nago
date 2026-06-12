// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xstrings"
)

// Compaction defaults. They are deliberately conservative so the default behaviour is safe; tune them via
// [SummaryCompactorConfig] for specific workloads.
const (
	// DefaultKeepLastN is the number of most recent messages kept verbatim by [NewSummaryCompactor]. Older
	// messages are folded into a single summary message.
	DefaultKeepLastN = 6

	// DefaultMaxSummaryTokens caps the output tokens of the summarization completion, bounding the size of
	// the produced summary.
	DefaultMaxSummaryTokens = 2048

	// defaultSummaryPrompt instructs the model to produce a dense, information-preserving summary.
	defaultSummaryPrompt = "You are compacting a conversation to fit a limited context window. " +
		"Summarize the following transcript as densely as possible while preserving all facts, decisions, " +
		"open questions, user preferences, tool results and any information required to continue the task. " +
		"Prefer bullet points over prose. Do not add commentary, only the summary."

	// minChunkRunes is the smallest text size the summarizer attempts before it falls back to a plain
	// rune-safe truncation. It also bounds the recursion of the divide-and-conquer summarizer.
	minChunkRunes = 2000

	// maxSummaryDepth bounds the divide-and-conquer recursion when even a single chunk overflows the window.
	maxSummaryDepth = 8
)

// SummaryCompactorConfig configures [NewSummaryCompactor].
type SummaryCompactorConfig struct {
	// KeepLastN is the number of most recent messages to keep verbatim. Zero means [DefaultKeepLastN].
	KeepLastN int

	// MaxSummaryTokens caps the generated summary length. Zero means [DefaultMaxSummaryTokens].
	MaxSummaryTokens int

	// Prompt overrides the summarization system prompt. Empty means a sensible default.
	Prompt string
}

// NewSummaryCompactor returns a [Compactor] that shrinks the history by summarizing the older messages into a
// single synthetic message while keeping the most recent ones verbatim. The summary is produced by issuing a
// completion request against the same model as the in-flight request (opts.Model).
//
// The strategy is robust and always terminates, regardless of the (unknown) context window size:
//
//   - It keeps the last N messages verbatim and folds everything before into one summary message. The split
//     point is moved so it never separates an assistant tool_use from its matching tool_result, keeping the
//     retained tail self-consistent.
//   - The older messages are rendered into a plain-text transcript and summarized via the model. If that
//     summarization request itself overflows the context window, the transcript is recursively split in half
//     (on rune boundaries) and summarized piece-wise (divide and conquer).
//   - As a final, guaranteed-to-succeed fallback, an oversized leaf chunk is truncated rune-safely via
//     [xstrings.EllipsisEnd] without calling the model at all.
//
// Because the produced summary is bounded by MaxSummaryTokens (or the truncation budget), the result is
// strictly smaller than the input whenever there is anything to compact, which guarantees progress for the
// retry loop in [Run].
func NewSummaryCompactor(cfg SummaryCompactorConfig) Compactor {
	keepLastN := cfg.KeepLastN
	if keepLastN <= 0 {
		keepLastN = DefaultKeepLastN
	}

	maxSummaryTokens := cfg.MaxSummaryTokens
	if maxSummaryTokens <= 0 {
		maxSummaryTokens = DefaultMaxSummaryTokens
	}

	prompt := cfg.Prompt
	if prompt == "" {
		prompt = defaultSummaryPrompt
	}

	return func(subject auth.Subject, c Completions, opts Options, history []Message) ([]Message, error) {
		if len(history) == 0 {
			return history, nil
		}

		start := splitIndex(history, keepLastN)
		prefix := history[:start]
		tail := history[start:]

		if len(prefix) == 0 {
			// Nothing older to summarize: fall back to rune-truncating the single largest message so the
			// request still shrinks and the run can make progress.
			return truncateLargest(history), nil
		}

		transcript := renderTranscript(prefix)
		summary := summarizeText(subject, c, opts, prompt, maxSummaryTokens, transcript, 0)

		summaryMsg := Message{
			Role:    User,
			Content: []Content{Text{Text: "[Summary of earlier conversation]\n" + summary}},
		}

		out := make([]Message, 0, len(tail)+1)
		out = append(out, summaryMsg)
		out = append(out, tail...)
		return out, nil
	}
}

// splitIndex computes the index of the first message to keep verbatim. It keeps up to keepLastN trailing
// messages and summarizes everything older. When the whole history already fits within the keep window the
// returned index is 0 (empty prefix), signalling the caller to fall back to truncation instead of
// summarizing the most recent messages. It never starts the retained tail on a message carrying tool_result
// blocks, which would dangle without their originating tool_use in the summarized prefix.
func splitIndex(history []Message, keepLastN int) int {
	n := len(history)

	start := n - keepLastN
	if start < 0 {
		start = 0
	}

	// Move the boundary forward while the first kept message would be a dangling tool_result.
	for start < n && hasToolResult(history[start]) {
		start++
	}

	return start
}

// summarizeText asks the model to summarize transcript. If the request overflows the context window it
// recursively splits the transcript in half (on rune boundaries) and summarizes both halves, joining the
// partial summaries. As a guaranteed fallback it rune-truncates the text without calling the model.
func summarizeText(subject auth.Subject, c Completions, opts Options, prompt string, maxSummaryTokens int, transcript string, depth int) string {
	res, err := c.Complete(subject, Options{
		Model:     opts.Model,
		System:    prompt,
		MaxTokens: maxSummaryTokens,
		Messages: []Message{
			{Role: User, Content: []Content{Text{Text: transcript}}},
		},
	})
	if err == nil {
		if s := strings.TrimSpace(extractText(res.Message)); s != "" {
			return s
		}
		// The model returned nothing usable; degrade to a rune-safe excerpt.
		return xstrings.EllipsisEnd(transcript, minChunkRunes)
	}

	if !errors.Is(err, ContextWindowExceeded) {
		// A non-overflow error (rate limit, transport, ...) cannot be fixed by splitting; degrade gracefully
		// to a rune-safe excerpt so the overall compaction still makes progress.
		return xstrings.EllipsisEnd(transcript, minChunkRunes)
	}

	// The transcript itself does not fit. Stop recursing once it is small enough or too deep and fall back to
	// a plain rune-safe truncation that cannot fail.
	if depth >= maxSummaryDepth || utf8.RuneCountInString(transcript) <= minChunkRunes {
		return xstrings.EllipsisEnd(transcript, minChunkRunes)
	}

	left, right := splitRunes(transcript)
	ls := summarizeText(subject, c, opts, prompt, maxSummaryTokens, left, depth+1)
	rs := summarizeText(subject, c, opts, prompt, maxSummaryTokens, right, depth+1)
	return ls + "\n" + rs
}

// splitRunes splits s into two halves on a rune boundary, never cutting a multi-byte codepoint.
func splitRunes(s string) (string, string) {
	runes := []rune(s)
	mid := len(runes) / 2
	return string(runes[:mid]), string(runes[mid:])
}

// renderTranscript turns a slice of messages into a plain-text transcript. Rendering to text (instead of
// forwarding the structured messages) sidesteps any tool_use/tool_result pairing constraints of the
// providers when the prefix is summarized.
func renderTranscript(msgs []Message) string {
	var sb strings.Builder
	for _, m := range msgs {
		sb.WriteString(string(m.Role))
		sb.WriteString(":\n")
		for _, content := range m.Content {
			switch c := content.(type) {
			case Text:
				sb.WriteString(c.Text)
			case Thinking:
				sb.WriteString("[thinking] ")
				sb.WriteString(c.Text)
			case ToolCall:
				sb.WriteString(fmt.Sprintf("[tool call %s %s]", c.Name, string(c.Arguments)))
			case ToolResult:
				sb.WriteString("[tool result")
				if c.IsError {
					sb.WriteString(" error")
				}
				sb.WriteString("] ")
				sb.WriteString(extractText(Message{Content: c.Content}))
			case Media:
				sb.WriteString(fmt.Sprintf("[media %s]", c.MimeType))
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// extractText concatenates all Text blocks of a message, ignoring non-text content.
func extractText(m Message) string {
	var sb strings.Builder
	for _, content := range m.Content {
		if t, ok := content.(Text); ok {
			sb.WriteString(t.Text)
		}
	}
	return sb.String()
}

// hasToolResult reports whether the message carries any tool_result content block.
func hasToolResult(m Message) bool {
	for _, content := range m.Content {
		if _, ok := content.(ToolResult); ok {
			return true
		}
	}
	return false
}

// truncateLargest returns a copy of history where the longest text-bearing message is rune-truncated to half
// its size. It is the last-resort path when there is no older prefix to summarize but the request still does
// not fit. It always shrinks the history, guaranteeing progress.
func truncateLargest(history []Message) []Message {
	out := make([]Message, len(history))
	copy(out, history)

	idx, maxLen := -1, 0
	for i, m := range out {
		if n := runeLenMsg(m); n > maxLen {
			idx, maxLen = i, n
		}
	}
	if idx < 0 || maxLen == 0 {
		return out
	}

	target := maxLen / 2
	if target < minChunkRunes {
		target = minChunkRunes
	}

	m := out[idx]
	newContent := make([]Content, len(m.Content))
	copy(newContent, m.Content)
	for i, content := range newContent {
		if t, ok := content.(Text); ok {
			newContent[i] = Text{Text: xstrings.EllipsisEnd(t.Text, target)}
		}
	}
	out[idx] = Message{Role: m.Role, Content: newContent}
	return out
}

// runeLen returns the total rune count across all messages, the metric [Run] uses to verify that a
// compaction strictly shrank the history.
func runeLen(history []Message) int {
	total := 0
	for _, m := range history {
		total += runeLenMsg(m)
	}
	return total
}

// runeLenMsg returns the rune count of the textual content of a single message.
func runeLenMsg(m Message) int {
	total := 0
	for _, content := range m.Content {
		switch c := content.(type) {
		case Text:
			total += utf8.RuneCountInString(c.Text)
		case Thinking:
			total += utf8.RuneCountInString(c.Text)
		case ToolCall:
			total += utf8.RuneCountInString(c.Name) + len(c.Arguments)
		case ToolResult:
			total += runeLenMsg(Message{Content: c.Content})
		}
	}
	return total
}


