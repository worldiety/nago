// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/drive"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xsync"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dropdown"
	"go.wdy.de/nago/presentation/ui/markdown"
)

// driveToolsChat demonstrates the LLM-driven file access flow: the model itself decides — via tool calls —
// which file from the nago drive it wants to look at. Two tools are offered:
//
//   - list_drive_files: a plain (text-result) tool that walks the user's drives and returns the available
//     files with their ids, names, mime types and sizes, so the model can pick one.
//   - open_drive_file:  a file-providing tool ([completion.NewOpenFileTool]). When the model calls it with a
//     file id, [completion.Run] reads the file from the drive, uploads it to the provider and attaches it to
//     the conversation as a Media block (referenced by file id). The model then sees the actual file content
//     (e.g. the pages of a PDF) and can answer questions about it.
//
// The heavy lifting (upload + Media injection on the correct user turn) is handled by the completion loop; the
// example only wires drive.Get/WalkDir to the tools and provides a FileUploader backed by provider.Files.
func driveToolsChat(wnd core.Window, uc ai.UseCases, drives drive.UseCases) core.View {
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
		return "Liste die Dateien im Drive auf, öffne die erste PDF und fasse ihren Inhalt zusammen."
	})
	answer := core.AutoState[string](wnd)
	trace := core.AutoState[string](wnd)
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
		trace.Set("")
	})

	providerOptions := make([]dropdown.Option[provider.ID], 0, len(entries))
	for _, e := range entries {
		providerOptions = append(providerOptions, dropdown.Option[provider.ID]{
			Value: e.prov.Identity(),
			Label: e.prov.Name(),
		})
	}

	submit := func() {
		question := strings.TrimSpace(prompt.Get())
		if question == "" || busy.Get() {
			return
		}

		busy.Set(true)
		answer.Set("")
		trace.Set("")

		subject := wnd.Subject()

		xsync.Go(func() error {
			res, history, err := completion.Run(subject, comps, completion.RunOptions{
				Options: completion.Options{
					Model:     selectedModel.Get(),
					System:    "You are a helpful assistant with access to a file drive. Use list_drive_files to discover files and open_drive_file to read a specific file before answering. Answer based on the file content.",
					MaxTokens: 1024,
					Messages: []completion.Message{
						{Role: completion.User, Content: []completion.Content{completion.Text{Text: question}}},
					},
				},
				Tools:        driveTools(subject, drives),
				FileUploader: driveFileUploader(files),
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
				trace.Set(renderTrace(history))
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

	return ui.VStack(
		ui.Text(fmt.Sprintf("LLM liest Drive-Dateien – %s (%s)", prov.Name(), selectedModel.Get())).Font(ui.Title),
		ui.Text("Das Modell entscheidet selbst per Tool-Call, welche Drive-Datei es öffnet. open_drive_file lädt die Datei hoch und die Loop schleust sie als Media-Block in die Konversation ein."),

		dropdown.Dropdown("Provider", providerOptions, selectedProvider.Get()).
			InputValue(selectedProvider).
			Disabled(busy.Get()).
			Frame(ui.Frame{}.FullWidth()),

		ui.TextField("Deine Eingabe", prompt.Get()).
			InputValue(prompt).
			Lines(3).
			FullWidth().
			Disabled(busy.Get()),

		ui.PrimaryButton(submit).
			Title("completion.Run mit Drive-Tools starten").
			Enabled(!busy.Get()),

		ui.If(busy.Get(), ui.Text("… das Modell durchsucht das Drive und liest Dateien")),

		ui.If(answer.Get() != "", ui.VStack(
			ui.Text("Antwort").Font(ui.SubTitle),
			markdown.RichText(answer.Get()),
		).Alignment(ui.Leading).
			FullWidth().
			BackgroundColor(ui.M2).
			Border(ui.Border{}.Radius(ui.L8)).
			Padding(ui.Padding{}.All(ui.L16))),

		ui.If(trace.Get() != "", ui.VStack(
			ui.Text("Message-Trace (Tool-Calls + eingeschleuste Media)").Font(ui.SubTitle),
			ui.CodeEditor(trace.Get()).Language("text").FullWidth(),
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

// driveFileUploader adapts a provider Files capability into a completion.FileUploader: it uploads the opened
// file's bytes and returns the provider-native file id under which the completion loop references it.
func driveFileUploader(files provider.Files) completion.FileUploader {
	return func(subject auth.Subject, f completion.OpenedFile) (file.ID, error) {
		uploaded, err := files.Put(subject, file.CreateOptions{
			Name:     f.Name,
			MimeType: f.MimeType,
			Purpose:  file.PurposeUserData,
			Open:     f.Open,
		})
		if err != nil {
			return "", err
		}
		return uploaded.ID, nil
	}
}

// driveTools builds the two drive-backed tools offered to the model. Both are bound to the current subject so
// the drive's per-file permissions are enforced.
func driveTools(subject auth.Subject, drives drive.UseCases) []completion.Tool {
	type listIn struct{}
	type driveFile struct {
		FID  string `json:"fid"`
		Name string `json:"name"`
		Mime string `json:"mime,omitempty"`
		Size int64  `json:"size,omitempty"`
	}
	type listOut struct {
		Files []driveFile `json:"files"`
	}

	list := completion.NewTool("list_drive_files",
		"lists the files available in the drive with their id, name, mime type and size",
		func(in listIn) (listOut, error) {
			var out listOut
			for d, err := range drives.ReadDrives(subject, subject.ID()) {
				if err != nil {
					return listOut{}, err
				}

				werr := drives.WalkDir(subject, d.Root, func(fid drive.FID, f drive.File, err error) error {
					if err != nil {
						return nil // skip unreadable entries
					}
					if f.IsDir() {
						return nil
					}
					mime := ""
					if f.FileInfo.IsSome() {
						mime = f.FileInfo.Unwrap().MimeType
					}
					out.Files = append(out.Files, driveFile{
						FID:  string(fid),
						Name: f.Name(),
						Mime: mime,
						Size: f.Size(),
					})
					return nil
				})
				if werr != nil {
					return listOut{}, werr
				}
			}
			return out, nil
		})

	type openFileIn struct {
		FID string `json:"fid" desc:"the id of the drive file to open, as returned by list_drive_files"`
	}

	open := completion.NewOpenFileTool("open_drive_file",
		"opens a drive file by its id and attaches its content to the conversation so it can be inspected",
		func(in openFileIn) (completion.OpenedFile, error) {
			optFile, err := drives.Get(subject, drive.FID(in.FID), "")
			if err != nil {
				return completion.OpenedFile{}, err
			}
			if optFile.IsNone() {
				return completion.OpenedFile{}, fmt.Errorf("no such file: %s", in.FID)
			}

			f := optFile.Unwrap()
			mimeStr, _ := f.MimeType()
			mime := driveMimeToFileType(mimeStr, f.Name())
			if mime == file.Binary {
				return completion.OpenedFile{}, fmt.Errorf("unsupported file type %q for %q; only images and PDFs can be attached", mimeStr, f.Name())
			}

			return completion.OpenedFile{
				Name:     f.Name(),
				MimeType: mime,
				Open:     f.Open,
			}, nil
		})

	return []completion.Tool{list, open}
}

// driveMimeToFileType maps a drive-reported mime string (falling back to the filename extension) to a
// [file.Type] the AI provider supports as an attachment. Unsupported types map to [file.Binary].
func driveMimeToFileType(mime, name string) file.Type {
	switch file.Type(mime) {
	case file.PNG, file.JPEG, file.GIF, file.PDF:
		return file.Type(mime)
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
		return file.Binary
	}
}
