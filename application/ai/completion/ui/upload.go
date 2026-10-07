// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

// maxUploadBytes caps the size of a single user-attached file.
const maxUploadBytes = 32 * 1024 * 1024

// maxInlineUploadTextBytes caps an inlined text attachment, like the text a file-providing tool injects, so that a
// single attachment cannot blow the context window of the model.
const maxInlineUploadTextBytes = 256 * 1024

// Attachment is a file the user attached to a message, see [ChatOptions.OnAttach].
type Attachment struct {
	Name string
	Mime file.Type
	Data []byte
}

// OnAttach takes over an attachment before the built-in handling, e.g. a workbook which the application stores
// itself and its tools work on. It runs on the background goroutine of the submit. With handled, content becomes
// part of the user turn, e.g. a text block telling the model the id of the stored file. Without, the built-in
// handling applies, which accepts images, PDFs and text only. An error is shown to the user and aborts the turn.
type OnAttach func(subject auth.Subject, attachment Attachment) (content []completion.Content, handled bool, err error)

// stagedFile is a file the user picked but has not sent yet. The raw bytes are held in memory until the next
// submit turns them into message content.
type stagedFile struct {
	Name string
	Mime file.Type
	Data []byte
}

// uploadButton renders the "attach file" button. Picked files are read into memory and appended to the staged
// state so they can be shown as chips and attached to the next message. It is only wired when
// [ChatOptions.FileUpload] is set and the provider exposes a Files capability.
func uploadButton(wnd core.Window, staged *core.State[[]stagedFile], disabled bool) core.View {
	pick := func() {
		wnd.ImportFiles(core.ImportFilesOptions{
			Multiple: true,
			MaxBytes: maxUploadBytes,
			OnCompletion: func(fs []core.File) {
				for _, f := range fs {
					r, err := f.Open()
					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}
					data, err := io.ReadAll(r)
					_ = r.Close()
					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					mimeStr, _ := f.MimeType()
					sf := stagedFile{Name: f.Name(), Mime: detectUploadMime(f.Name(), mimeStr), Data: data}
					wnd.Post(func() {
						staged.Set(append(staged.Get(), sf))
					})
				}
			},
		})
	}

	return ui.SecondaryButton(pick).
		PreIcon(icons.Upload).
		AccessibilityLabel("Datei anhängen").
		Title("Datei").
		Enabled(!disabled)
}

// stagedChips renders the currently staged (not yet sent) files as removable chips.
func stagedChips(staged *core.State[[]stagedFile], disabled bool) core.View {
	files := staged.Get()
	if len(files) == 0 {
		return nil
	}

	chips := make([]core.View, 0, len(files))
	for i, f := range files {
		i := i
		chips = append(chips, ui.HStack(
			ui.Text(fmt.Sprintf("%s (%d B)", f.Name, len(f.Data))).Font(ui.Small),
			ui.TertiaryButton(func() {
				cur := staged.Get()
				if i < 0 || i >= len(cur) {
					return
				}
				staged.Set(append(append([]stagedFile{}, cur[:i]...), cur[i+1:]...))
			}).PreIcon(icons.Close).AccessibilityLabel("Entfernen").Enabled(!disabled),
		).Gap(ui.L4).Alignment(ui.Center).
			BackgroundColor(ui.M3).
			Border(ui.Border{}.Radius(ui.L8)).
			Padding(ui.Padding{}.Horizontal(ui.L8).Vertical(ui.L4)))
	}

	return ui.HStack(chips...).Gap(ui.L4).FullWidth().Alignment(ui.Leading)
}

// buildUploadContent turns the staged files into leading message content blocks for the next user turn. Images
// and PDFs are uploaded to the provider and referenced by file id (bytes travel once); text files are inlined
// so the model can read them directly. Unsupported binary files are rejected with an error. It runs on the
// background submit goroutine.
func buildUploadContent(subject auth.Subject, files provider.Files, owner file.Owner, staged []stagedFile, onAttach OnAttach) ([]completion.Content, error) {
	var content []completion.Content
	for _, sf := range staged {
		if onAttach != nil {
			taken, handled, err := onAttach(subject, Attachment{Name: sf.Name, Mime: sf.Mime, Data: sf.Data})
			if err != nil {
				return nil, err
			}

			if handled {
				content = append(content, taken...)
				continue
			}
		}

		if isImageMime(sf.Mime) || sf.Mime == file.PDF {
			if files == nil {
				return nil, fmt.Errorf("Datei %q kann nicht angehängt werden: Provider unterstützt keine Datei-Uploads", sf.Name)
			}
			data := sf.Data
			uploaded, err := files.Put(subject, file.CreateOptions{
				Name:     sf.Name,
				MimeType: sf.Mime,
				Purpose:  file.PurposeUserData,
				Owner:    owner,
				Open: func() (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader(data)), nil
				},
			})
			if err != nil {
				return nil, fmt.Errorf("Upload von %q: %w", sf.Name, err)
			}
			content = append(content, completion.Media{
				MimeType: sf.Mime,
				Source:   completion.Source{FileID: option.Some(uploaded.ID)},
			})
			continue
		}

		// Text files are inlined directly (no upload needed, works with every provider).
		if file.IsText(sf.Mime) || isProbablyText(sf.Data) {
			data := truncateUTF8(sf.Data, maxInlineUploadTextBytes)
			var sb strings.Builder
			sb.WriteString("--- Datei: ")
			sb.WriteString(sf.Name)
			sb.WriteString(" ---\n")
			sb.Write(data)
			if len(data) < len(sf.Data) {
				fmt.Fprintf(&sb, "\n--- gekürzt: nur die ersten %d KB von %d KB ---", len(data)/1024, len(sf.Data)/1024)
			}
			content = append(content, completion.Text{Text: sb.String()})
			continue
		}

		return nil, fmt.Errorf("nicht unterstützter Dateityp %q für %q – bitte Bild, PDF oder Textdatei anhängen", sf.Mime, sf.Name)
	}

	return content, nil
}

// isImageMime mirrors the provider's image classification for the supported image types.
// uploadOwner returns the owner of the files a chat uploads: its session, or the transient owner for a chat
// without History, whose files are deleted a day later.
func uploadOwner(sid session.ID) file.Owner {
	if sid == "" {
		return file.TransientOwner
	}

	return file.Owner(sid)
}

// truncateUTF8 cuts data to at most limit bytes, without splitting a rune.
func truncateUTF8(data []byte, limit int) []byte {
	if len(data) <= limit {
		return data
	}

	n := limit
	for n > 0 && !utf8.RuneStart(data[n]) {
		n--
	}

	return data[:n]
}

func isImageMime(t file.Type) bool {
	switch t {
	case file.PNG, file.JPEG, file.GIF:
		return true
	default:
		return false
	}
}

// detectUploadMime maps a browser-reported mime string (or the filename extension as fallback) to a
// [file.Type], preferring the canonical image/PDF types the provider can attach.
func detectUploadMime(name, browserMime string) file.Type {
	if i := strings.IndexByte(browserMime, ';'); i >= 0 {
		browserMime = strings.TrimSpace(browserMime[:i])
	}

	switch file.Type(browserMime) {
	case file.PNG, file.JPEG, file.GIF, file.PDF:
		return file.Type(browserMime)
	}
	if browserMime != "" && file.IsText(file.Type(browserMime)) {
		return file.Type(browserMime)
	}

	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return file.PNG
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return file.JPEG
	case strings.HasSuffix(lower, ".gif"):
		return file.GIF
	case strings.HasSuffix(lower, ".pdf"):
		return file.PDF
	}

	if browserMime != "" {
		return file.Type(browserMime)
	}
	return file.Binary
}

// isProbablyText reports whether data looks like UTF-8 text (no NUL bytes, valid rune sequence) so it can be
// safely inlined into a prompt.
func isProbablyText(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return false
	}
	return utf8.Valid(data)
}
