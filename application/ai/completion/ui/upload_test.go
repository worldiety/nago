// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

// A text attachment is inlined like the text of a file-providing tool: capped, with a note, and never split
// within a character.
func TestTextAttachmentIsCapped(t *testing.T) {
	big := strings.Repeat("ä", maxInlineUploadTextBytes) // two bytes each
	content, err := buildUploadContent(user.SU(), nil, file.TransientOwner, []stagedFile{{Name: "big.txt", Mime: file.Type("text/plain"), Data: []byte(big)}}, nil)
	if err != nil {
		t.Fatal(err)
	}

	text := content[0].(completion.Text).Text
	if len(text) > maxInlineUploadTextBytes+200 || !strings.Contains(text, "gekürzt") || !utf8.ValidString(text) {
		t.Fatalf("expected a capped text of valid UTF-8 with a note, got %d bytes", len(text))
	}

	small, err := buildUploadContent(user.SU(), nil, file.TransientOwner, []stagedFile{{Name: "a.txt", Mime: file.Type("text/plain"), Data: []byte("hello")}}, nil)
	if err != nil || strings.Contains(small[0].(completion.Text).Text, "gekürzt") {
		t.Fatalf("a small text must stay complete: %v %v", small, err)
	}
}

// The application takes over attachments nago cannot send, e.g. a workbook, and declines others, which keep the
// built-in handling.
func TestOnAttachTakesOverAttachments(t *testing.T) {
	xlsx := file.Type("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	onAttach := func(subject auth.Subject, att Attachment) ([]completion.Content, bool, error) {
		switch {
		case att.Mime == xlsx:
			return []completion.Content{completion.Text{Text: "workbook " + att.Name + " stored"}}, true, nil
		case att.Name == "virus.exe":
			return nil, false, errors.New("not allowed")
		default:
			return nil, false, nil
		}
	}

	content, err := buildUploadContent(user.SU(), nil, file.TransientOwner, []stagedFile{
		{Name: "inventory.xlsx", Mime: xlsx, Data: []byte{0x50, 0x4b, 0x03, 0x04, 0x00}},
		{Name: "notes.txt", Mime: file.Type("text/plain"), Data: []byte("hello")},
	}, onAttach)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) != 2 || content[0].(completion.Text).Text != "workbook inventory.xlsx stored" || !strings.Contains(content[1].(completion.Text).Text, "hello") {
		t.Fatalf("unexpected content %#v", content)
	}

	if _, err := buildUploadContent(user.SU(), nil, file.TransientOwner, []stagedFile{{Name: "virus.exe", Mime: file.Type("application/octet-stream"), Data: []byte{0}}}, onAttach); err == nil || err.Error() != "not allowed" {
		t.Fatalf("expected the error of the application, got %v", err)
	}

	// without OnAttach, a workbook is still rejected
	if _, err := buildUploadContent(user.SU(), nil, file.TransientOwner, []stagedFile{{Name: "inventory.xlsx", Mime: xlsx, Data: []byte{0x50, 0x4b, 0x03, 0x04, 0x00}}}, nil); err == nil {
		t.Fatal("expected an unsupported workbook to be rejected without OnAttach")
	}
}

// capturingFiles records the options of each upload.
type capturingFiles struct {
	provider.Files
	puts []file.CreateOptions
}

func (f *capturingFiles) Put(_ auth.Subject, opts file.CreateOptions) (file.File, error) {
	f.puts = append(f.puts, opts)
	return file.File{ID: "file-1"}, nil
}

// The attachments of a chat are uploaded on behalf of its session, so that they are deleted at the provider with
// the session. A chat without History has no session, its files are transient.
func TestAttachmentsAreOwnedByTheSession(t *testing.T) {
	files := &capturingFiles{}
	pdf := []stagedFile{{Name: "a.pdf", Mime: file.PDF, Data: []byte("%PDF")}}

	if _, err := buildUploadContent(user.SU(), files, uploadOwner("session-1"), pdf, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := buildUploadContent(user.SU(), files, uploadOwner(""), pdf, nil); err != nil {
		t.Fatal(err)
	}

	if len(files.puts) != 2 || files.puts[0].Owner != "session-1" || files.puts[1].Owner != file.TransientOwner {
		t.Fatalf("unexpected owners %+v", files.puts)
	}

	// an application which takes over an attachment learns the conversation as well
	var owner file.Owner
	onAttach := func(_ auth.Subject, att Attachment) ([]completion.Content, bool, error) {
		owner = att.Owner
		return []completion.Content{completion.Text{Text: "taken"}}, true, nil
	}
	if _, err := buildUploadContent(user.SU(), nil, uploadOwner("session-1"), pdf, onAttach); err != nil || owner != "session-1" {
		t.Fatalf("expected the session as owner of the taken attachment, got %q %v", owner, err)
	}
}

// A long file name is shortened in the middle, so that a chip keeps its extension visible.
func TestShortName(t *testing.T) {
	if got := shortName("probe.xlsx", 32); got != "probe.xlsx" {
		t.Fatalf("a short name must stay, got %q", got)
	}

	got := shortName("inventur-lager-hamburg-2026-oktober-final.xlsx", 32)
	if utf8.RuneCountInString(got) != 32 || !strings.HasSuffix(got, "final.xlsx") || !strings.Contains(got, "…") {
		t.Fatalf("unexpected short name %q", got)
	}
}
