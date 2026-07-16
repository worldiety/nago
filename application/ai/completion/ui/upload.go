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
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

// maxUploadBytes caps the size of a single user-attached file.
const maxUploadBytes = 32 * 1024 * 1024

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
func buildUploadContent(subject auth.Subject, files provider.Files, staged []stagedFile) ([]completion.Content, error) {
	var content []completion.Content
	for _, sf := range staged {
		if isImageMime(sf.Mime) || sf.Mime == file.PDF {
			if files == nil {
				return nil, fmt.Errorf("Datei %q kann nicht angehängt werden: Provider unterstützt keine Datei-Uploads", sf.Name)
			}
			data := sf.Data
			uploaded, err := files.Put(subject, file.CreateOptions{
				Name:     sf.Name,
				MimeType: sf.Mime,
				Purpose:  file.PurposeUserData,
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
			var sb strings.Builder
			sb.WriteString("--- Datei: ")
			sb.WriteString(sf.Name)
			sb.WriteString(" ---\n")
			sb.Write(sf.Data)
			content = append(content, completion.Text{Text: sb.String()})
			continue
		}

		return nil, fmt.Errorf("nicht unterstützter Dateityp %q für %q – bitte Bild, PDF oder Textdatei anhängen", sf.Mime, sf.Name)
	}

	return content, nil
}

// isImageMime mirrors the provider's image classification for the supported image types.
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
