// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/auth"
)

const (
	// titleTimeout bounds the completion, which summarizes a conversation into its title.
	titleTimeout = 30 * time.Second

	// titleMinWords is the number of words the user must have written, before a conversation has a topic.
	// A greeting like "Hallo" has none yet.
	titleMinWords = 4

	// titleTranscriptRunes bounds the part of the conversation, which is summarized.
	titleTranscriptRunes = 2000

	// titleMaxRunes bounds the title, in case the model ignores the instruction.
	titleMaxRunes = 60

	titleInstruction = "Du erzeugst den Titel einer Unterhaltung für eine Übersicht. Formuliere das Thema in höchstens sechs Wörtern, in der Sprache der Unterhaltung. Antworte nur mit dem Titel, ohne Anführungszeichen und ohne Satzzeichen am Ende."
)

// titling holds the sessions, for which a title is being generated, so that a second turn does not start another
// generation.
var titling sync.Map

// autoTitle summarizes the conversation into its title once, after its first meaningful turn. A session which has
// a title is never summarized again, so neither a later turn nor a title given by the user changes it. A title
// equal to legacyTitle is the title of the chat panel, which former versions gave every conversation, and is
// replaced. The completion runs in the background with the model of the session and is not part of the
// conversation. If it fails, the next turn tries again.
func autoTitle(subject auth.Subject, comps completion.Completions, sessions session.UseCases, s session.Session, legacyTitle string) {
	if comps == nil || sessions.Rename == nil || sessions.FindByID == nil || !untitled(s, legacyTitle) || !meaningful(s.Messages) {
		return
	}

	if _, busy := titling.LoadOrStore(s.ID, struct{}{}); busy {
		return
	}

	go func() {
		defer titling.Delete(s.ID)

		ctx, cancel := context.WithTimeout(context.Background(), titleTimeout)
		defer cancel()

		title, err := generateTitle(ctx, subject, comps, s)
		if err != nil {
			slog.Warn("cannot summarize the conversation into a title", "session", s.ID, "err", err)
			return
		}

		if title == "" {
			return
		}

		// the user may have given a title in the meantime
		optSession, err := sessions.FindByID(subject, s.ID)
		if err != nil || optSession.IsNone() || !untitled(optSession.Unwrap(), legacyTitle) {
			return
		}

		if err := sessions.Rename(subject, s.ID, title); err != nil {
			slog.Warn("cannot store the title of the conversation", "session", s.ID, "err", err)
		}
	}()
}

func untitled(s session.Session, legacyTitle string) bool {
	return s.Title == "" || s.Title == legacyTitle
}

// meaningful tells, whether the conversation has a topic: the user has written enough or attached something, and
// the assistant has answered with text.
func meaningful(messages []completion.Message) bool {
	words := 0
	attached := false
	answered := false
	for _, msg := range messages {
		for _, c := range msg.Content {
			switch c := c.(type) {
			case completion.Text:
				switch msg.Role {
				case completion.User:
					words += len(strings.Fields(c.Text))
				case completion.Assistant:
					answered = answered || strings.TrimSpace(c.Text) != ""
				}
			case completion.ToolResult:
				// tool results are sent in user messages, but the user has not written them
			default:
				attached = attached || msg.Role == completion.User
			}
		}
	}

	return answered && (words >= titleMinWords || attached)
}

func generateTitle(ctx context.Context, subject auth.Subject, comps completion.Completions, s session.Session) (string, error) {
	res, err := comps.Complete(ctx, subject, completion.Options{
		Model:  s.Model,
		System: titleInstruction,
		Messages: []completion.Message{{
			Role:    completion.User,
			Content: []completion.Content{completion.Text{Text: transcript(s.Messages)}},
		}},
		MaxTokens: 64,
	})
	if err != nil {
		return "", err
	}

	var text strings.Builder
	for _, c := range res.Message.Content {
		if t, ok := c.(completion.Text); ok {
			text.WriteString(t.Text)
		}
	}

	return cleanTitle(text.String()), nil
}

// transcript renders the beginning of the text of the conversation, without tool calls and attachments.
func transcript(messages []completion.Message) string {
	var sb strings.Builder
	for _, msg := range messages {
		var speaker string
		switch msg.Role {
		case completion.User:
			speaker = "Nutzer"
		case completion.Assistant:
			speaker = "Assistent"
		default:
			continue
		}

		for _, c := range msg.Content {
			t, ok := c.(completion.Text)
			if !ok || strings.TrimSpace(t.Text) == "" {
				continue
			}

			sb.WriteString(speaker)
			sb.WriteString(": ")
			sb.WriteString(strings.TrimSpace(t.Text))
			sb.WriteString("\n\n")
		}

		if utf8.RuneCountInString(sb.String()) >= titleTranscriptRunes {
			break
		}
	}

	return truncateRunes(sb.String(), titleTranscriptRunes)
}

// cleanTitle takes the first line of the answer and removes what a model tends to add anyway.
func cleanTitle(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimLeft(strings.TrimSpace(line), "#*-> ")
		line = strings.TrimPrefix(line, "Titel:")
		// quotes and punctuation may be nested, e.g. a quoted title followed by a period
		for prev := ""; prev != line; {
			prev = line
			line = strings.TrimSpace(strings.TrimRight(strings.Trim(line, "\"'„“”«»*` "), ".:;!"))
		}

		if line != "" {
			return truncateRunes(line, titleMaxRunes)
		}
	}

	return ""
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}

	return strings.TrimSpace(string(r[:n-1])) + "…"
}
