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
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/drive"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

// driveToolsChat demonstrates the LLM-driven file access flow via the generic [uicompletion.Chat] component:
// the model itself decides — via tool calls — which file from the nago drive it wants to look at. Two tools
// are offered:
//
//   - list_drive_files: a plain (text-result) tool that walks the user's drives and returns the available
//     files with their ids, names, mime types and sizes, so the model can pick one.
//   - open_drive_file:  a file-providing tool ([completion.NewOpenFileTool]). When the model calls it with a
//     file id, the completion loop reads the file from the drive, uploads it to the provider (wired via
//     ChatOptions.FileUpload) and attaches it to the conversation as a Media block. The model then sees the
//     actual file content (e.g. the pages of a PDF) and can answer questions about it.
//
// All the heavy lifting is handled by the component and the completion loop; the example only supplies the two
// drive-backed tools and turns FileUpload on.
func driveToolsChat(wnd core.Window, uc ai.UseCases, sessions session.UseCases, drives drive.UseCases) core.View {
	prov, comps, err := firstCompletionProvider(wnd.Subject(), uc, true)
	if err != nil {
		return alert.BannerError(err)
	}

	chat := uicompletion.Chat(wnd, uicompletion.ChatOptions{
		Sessions:    sessions,
		Completions: comps,
		Provider:    prov,
		Title:       "Drive-Tools",
		FileUpload:  true,
		Agents: []uicompletion.Agent{{
			SystemPrompt: "You are a helpful assistant with access to a file drive. Use list_drive_files to discover files and open_drive_file to read a specific file before answering. Answer based on the file content.",
			Tools: func(subject auth.Subject) []completion.Tool {
				return driveTools(subject, drives)
			},
		}},
	})

	return ui.VStack(
		ui.Text("LLM liest Drive-Dateien (uicompletion.Chat mit Tools + FileUpload)").Font(ui.Title),
		ui.Text("Das Modell entscheidet selbst per Tool-Call, welche Drive-Datei es öffnet. open_drive_file lädt die Datei hoch und die Loop schleust sie als Media-Block in die Konversation ein."),
		chat,
	).Alignment(ui.Leading).
		Gap(ui.L16).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L16))
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
		"opens a drive file by its id and makes its content available to the conversation. Text files (e.g. .md, .txt, .csv, .json, source code) are injected inline as text; images and PDFs are attached as media.",
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
				return completion.OpenedFile{}, fmt.Errorf("unsupported file type %q for %q; only text files, images and PDFs can be opened", mimeStr, f.Name())
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
// [file.Type] the AI chat can handle: text types are injected inline by the completion loop, while images and
// PDFs are attached as media. Unsupported binary types map to [file.Binary].
func driveMimeToFileType(mime, name string) file.Type {
	// defensively strip any RFC 6838 parameters (e.g. "text/plain; charset=binary") that older drive
	// records may still carry, so the comparisons below see the bare media type.
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}

	// canonical media types we attach as-is
	switch file.Type(mime) {
	case file.PNG, file.JPEG, file.GIF, file.PDF:
		return file.Type(mime)
	}

	// any text-ish mime (text/*, application/json, ...) is injected inline as text
	if file.IsText(file.Type(mime)) {
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
	case strings.HasSuffix(lower, ".md"), strings.HasSuffix(lower, ".markdown"):
		return file.Markdown
	case strings.HasSuffix(lower, ".csv"):
		return file.CSV
	case strings.HasSuffix(lower, ".json"):
		return file.JSON
	case strings.HasSuffix(lower, ".xml"):
		return file.XML
	case strings.HasSuffix(lower, ".txt"),
		strings.HasSuffix(lower, ".log"),
		strings.HasSuffix(lower, ".go"),
		strings.HasSuffix(lower, ".ts"),
		strings.HasSuffix(lower, ".js"),
		strings.HasSuffix(lower, ".py"),
		strings.HasSuffix(lower, ".java"),
		strings.HasSuffix(lower, ".rs"),
		strings.HasSuffix(lower, ".c"),
		strings.HasSuffix(lower, ".h"),
		strings.HasSuffix(lower, ".yaml"),
		strings.HasSuffix(lower, ".yml"),
		strings.HasSuffix(lower, ".toml"),
		strings.HasSuffix(lower, ".ini"),
		strings.HasSuffix(lower, ".sh"),
		strings.HasSuffix(lower, ".sql"),
		strings.HasSuffix(lower, ".html"),
		strings.HasSuffix(lower, ".htm"):
		return file.Text
	default:
		return file.Binary
	}
}
