// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xsync"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dropdown"
	"go.wdy.de/nago/presentation/ui/markdown"
)

// fileUploadChat demonstrates the file *upload* flow: the user picks a file, it is uploaded to the provider via
// its Files capability, and then attached to a user message as a [completion.Media] block referencing the
// uploaded file by its provider-native file id. The model receives the file plus the user's question and
// explains its contents — proving it actually read the file instead of guessing.
//
// This is the natural direction for file handling: bytes go to the provider once (via file id), and the
// reference stays small on every turn. It requires a provider that exposes BOTH stateless completions AND a
// Files capability (e.g. a configured Anthropic secret).
//
// Supported file kinds and how they are sent to Anthropic (see anthropic/conv.go):
//   - images (PNG/JPEG/GIF) -> image block sourced by file id
//   - PDF                   -> document block sourced by file id
//   - other (e.g. CSV/TXT)  -> the raw text is inlined into the prompt (Anthropic's document blocks only
//     accept PDF/plain-text file ids; to keep this example simple and robust we read small text files
//     directly instead of uploading them).
func fileUploadChat(wnd core.Window, uc ai.UseCases) core.View {
	type provEntry struct {
		prov  provider.Provider
		comps completion.Completions
		files provider.Files
	}

	var entries []provEntry
	for p, err := range uc.FindAllProvider(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		c := p.Completions()
		f := p.Files()
		if c.IsSome() && f.IsSome() {
			entries = append(entries, provEntry{prov: p, comps: c.Unwrap(), files: f.Unwrap()})
		}
	}

	if len(entries) == 0 {
		return alert.BannerError(fmt.Errorf("kein Provider mit Completions UND Files gefunden – bitte ein Anthropic-Secret konfigurieren"))
	}

	selectedProvider := core.AutoState[provider.ID](wnd).Init(func() provider.ID {
		return entries[0].prov.Identity()
	})

	current := entries[0]
	for _, e := range entries {
		if e.prov.Identity() == selectedProvider.Get() {
			current = e
			break
		}
	}
	prov := current.prov
	comps := current.comps
	files := current.files

	prompt := core.AutoState[string](wnd).Init(func() string {
		return "Fasse den Inhalt dieser Datei zusammen und nenne die drei wichtigsten Punkte."
	})
	// The picked file, staged in memory until the request is sent.
	pickedName := core.AutoState[string](wnd)
	pickedMime := core.AutoState[string](wnd)
	pickedBytes := core.AutoState[[]byte](wnd)

	answer := core.AutoState[string](wnd)
	usage := core.AutoState[string](wnd)
	busy := core.AutoState[bool](wnd)
	selectedModel := core.AutoState[model.ID](wnd).Init(func() model.ID {
		for m, err := range comps.Models(wnd.Subject()) {
			if err != nil {
				return ""
			}
			return m.ID
		}
		return ""
	})

	selectedProvider.Observe(func(newValue provider.ID) {
		first := model.ID("")
		for _, e := range entries {
			if e.prov.Identity() == newValue {
				for m, err := range e.comps.Models(wnd.Subject()) {
					if err == nil {
						first = m.ID
					}
					break
				}
				break
			}
		}
		selectedModel.Set(first)
		answer.Set("")
		usage.Set("")
	})

	providerOptions := make([]dropdown.Option[provider.ID], 0, len(entries))
	for _, e := range entries {
		providerOptions = append(providerOptions, dropdown.Option[provider.ID]{
			Value: e.prov.Identity(),
			Label: e.prov.Name(),
		})
	}

	pickFile := func() {
		wnd.ImportFiles(core.ImportFilesOptions{
			Multiple: false,
			MaxBytes: 16 * 1024 * 1024,
			OnCompletion: func(fs []core.File) {
				if len(fs) == 0 {
					return
				}
				f := fs[0]
				r, err := f.Open()
				if err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
				defer r.Close()

				data, err := io.ReadAll(r)
				if err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				mime, _ := f.MimeType()
				wnd.Post(func() {
					pickedName.Set(f.Name())
					pickedMime.Set(mime)
					pickedBytes.Set(data)
					answer.Set("")
					usage.Set("")
				})
			},
		})
	}

	submit := func() {
		question := strings.TrimSpace(prompt.Get())
		if question == "" || busy.Get() {
			return
		}
		if len(pickedBytes.Get()) == 0 {
			alert.ShowBannerError(wnd, fmt.Errorf("bitte zuerst eine Datei auswählen"))
			return
		}

		busy.Set(true)
		answer.Set("")
		usage.Set("")

		name := pickedName.Get()
		mime := detectMime(name, pickedMime.Get())
		data := pickedBytes.Get()

		xsync.Go(func() error {
			userContent, err := buildFileMessage(wnd.Subject(), files, name, mime, data, question)
			if err != nil {
				wnd.Post(func() {
					busy.Set(false)
					alert.ShowBannerError(wnd, err)
				})
				return nil
			}

			res, err := comps.Complete(wnd.Subject(), completion.Options{
				Model:     selectedModel.Get(),
				System:    "You are a helpful assistant. Answer strictly based on the attached file. If the file does not contain the answer, say so.",
				MaxTokens: 1024,
				Messages: []completion.Message{
					{Role: completion.User, Content: userContent},
				},
			})
			if err != nil {
				wnd.Post(func() {
					busy.Set(false)
					alert.ShowBannerError(wnd, err)
				})
				return nil
			}

			var sb strings.Builder
			for _, c := range res.Message.Content {
				if t, ok := c.(completion.Text); ok {
					sb.WriteString(t.Text)
				}
			}

			wnd.Post(func() {
				answer.Set(sb.String())
				usage.Set(renderUsage(res.Usage))
				busy.Set(false)
			})
			return nil
		}, func(err error) {
			if err != nil {
				wnd.Post(func() {
					busy.Set(false)
					alert.ShowBannerError(wnd, err)
				})
			}
		})
	}

	fileLabel := "keine Datei ausgewählt"
	if n := pickedName.Get(); n != "" {
		fileLabel = fmt.Sprintf("%s (%s, %d Bytes)", n, detectMime(n, pickedMime.Get()), len(pickedBytes.Get()))
	}

	return ui.VStack(
		ui.Text(fmt.Sprintf("Datei hochladen und verstehen lassen – %s (%s)", prov.Name(), selectedModel.Get())).Font(ui.Title),
		ui.Text("Wähle eine Datei (Bild, PDF oder Textdatei). Sie wird zum Provider hochgeladen und dem Modell zusammen mit deiner Frage übergeben. Das Modell antwortet ausschließlich anhand des Datei-Inhalts."),

		dropdown.Dropdown("Provider", providerOptions, selectedProvider.Get()).
			InputValue(selectedProvider).
			Disabled(busy.Get()).
			Frame(ui.Frame{}.FullWidth()),

		ui.HStack(
			ui.SecondaryButton(pickFile).Title("Datei auswählen").Enabled(!busy.Get()),
			ui.Text(fileLabel),
		).Gap(ui.L8).Alignment(ui.Leading),

		ui.TextField("Deine Frage zur Datei", prompt.Get()).
			InputValue(prompt).
			Lines(3).
			FullWidth().
			Disabled(busy.Get()),

		ui.PrimaryButton(submit).
			Title("Frage an das Modell senden").
			Enabled(!busy.Get()),

		ui.If(busy.Get(), ui.Text("… Datei wird hochgeladen und analysiert")),

		ui.If(answer.Get() != "", ui.VStack(
			ui.Text("Antwort").Font(ui.SubTitle),
			markdown.RichText(answer.Get()),
		).Alignment(ui.Leading).
			FullWidth().
			BackgroundColor(ui.M2).
			Border(ui.Border{}.Radius(ui.L8)).
			Padding(ui.Padding{}.All(ui.L16))),

		ui.If(usage.Get() != "", ui.VStack(
			ui.Text("Token-Usage").Font(ui.SubTitle),
			ui.CodeEditor(usage.Get()).Language("text").FullWidth(),
		).Alignment(ui.Leading).
			FullWidth().
			BackgroundColor(ui.M2).
			Border(ui.Border{}.Radius(ui.L8)).
			Padding(ui.Padding{}.All(ui.L16))),
	).Alignment(ui.Leading).
		Gap(ui.L16).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L16))
}

// buildFileMessage turns a picked file into the content blocks of a user message. Images and PDFs are uploaded
// to the provider and attached by file id (efficient: the bytes travel once). Small text files are inlined
// into the prompt, because Anthropic's document blocks do not accept arbitrary text file ids.
func buildFileMessage(subject auth.Subject, files provider.Files, name string, mime file.Type, data []byte, question string) ([]completion.Content, error) {
	if isImageMime(mime) || mime == file.PDF {
		f, err := files.Put(subject, file.CreateOptions{
			Name:     name,
			MimeType: mime,
			Purpose:  file.PurposeUserData,
			Open: func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(data)), nil
			},
		})
		if err != nil {
			return nil, fmt.Errorf("upload file: %w", err)
		}

		return []completion.Content{
			completion.Media{MimeType: mime, Source: completion.Source{FileID: option.Some(f.ID)}},
			completion.Text{Text: question},
		}, nil
	}

	// Fallback for text-like files: inline the content directly so the model can read it.
	if !isProbablyText(data) {
		return nil, fmt.Errorf("nicht unterstützter Dateityp %q – bitte ein Bild, PDF oder eine Textdatei wählen", mime)
	}

	var sb strings.Builder
	sb.WriteString(question)
	sb.WriteString("\n\n--- Datei: ")
	sb.WriteString(name)
	sb.WriteString(" ---\n")
	sb.Write(data)

	return []completion.Content{completion.Text{Text: sb.String()}}, nil
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

// detectMime maps the browser-reported mime type (or the filename extension as a fallback) to a [file.Type].
func detectMime(name, browserMime string) file.Type {
	switch file.Type(browserMime) {
	case file.PNG, file.JPEG, file.GIF, file.PDF:
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
	default:
		if browserMime != "" {
			return file.Type(browserMime)
		}
		return file.Binary
	}
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
