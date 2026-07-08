// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidrive

import (
	"io"
	"log/slog"
	"mime"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/drive"
	drivehttp "go.wdy.de/nago/application/drive/http"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/markdown"
	"go.wdy.de/nago/presentation/ui/video"
	"golang.org/x/text/language"
)

// maxTextPreviewBytes caps how much of a text/markdown file is loaded into memory for an inline preview so a
// huge file cannot exhaust server memory. Larger files fall back to the download option.
const maxTextPreviewBytes = 2 * 1024 * 1024 // 2 MiB

var (
	StrPreviewNoPreviewTitle = i18n.MustString("nago.drive.preview.no_preview_title", i18n.Values{language.English: "No preview available", language.German: "Keine Vorschau verfügbar"})
	StrPreviewNoPreviewDesc  = i18n.MustString("nago.drive.preview.no_preview_desc", i18n.Values{language.English: "No preview can be displayed for this file. Download it to view it in a suitable application.", language.German: "Für diese Datei kann keine Vorschau angezeigt werden. Laden Sie die Datei herunter, um sie in einer passenden Anwendung anzuzeigen."})
	StrPreviewDownload       = i18n.MustString("nago.drive.preview.download", i18n.Values{language.English: "Download", language.German: "Herunterladen"})
	StrPreviewClose          = i18n.MustString("nago.drive.preview.close", i18n.Values{language.English: "Close preview", language.German: "Vorschau schließen"})
)

// previewKind classifies how a file can be previewed inline.
type previewKind int

const (
	previewNone previewKind = iota
	previewImage
	previewVideo
	previewMarkdown
	previewText
)

// dialogPreview renders a full screen preview overlay for the file identified by previewFid. It dispatches on
// the file mime type / extension: images via ui.Image, videos via the video player, markdown via a rich text
// renderer and any other utf8 text via a read-only code editor. Everything else shows a prominent download
// action.
func (c TDrive) dialogPreview(wnd core.Window, uc drive.UseCases, presented *core.State[bool], previewFid *core.State[drive.FID]) core.View {
	if !presented.Get() || previewFid.Get() == "" {
		return nil
	}

	fid := previewFid.Get()

	optFile, err := c.stat(fid)
	if err != nil {
		alert.ShowBannerError(wnd, err)
		return nil
	}
	if optFile.IsNone() {
		return nil
	}

	file := optFile.Unwrap()

	title := file.Filename
	closeFn := func() {
		presented.Set(false)
	}

	body := c.previewContent(wnd, uc, file)

	header := ui.HStack(
		ui.Text(title).Font(ui.TitleMedium),
		ui.Spacer(),
		ui.TertiaryButton(func() {
			c.downloadFile(wnd, uc, file.ID)
		}).PreIcon(icons.Download).Title(StrPreviewDownload.Get(wnd)),
		ui.TertiaryButton(closeFn).PreIcon(icons.Close).AccessibilityLabel(StrPreviewClose.Get(wnd)),
	).
		Alignment(ui.Center).
		Gap(ui.L8).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L16))

	content := ui.VStack(
		header,
		ui.HLine(),
		ui.ScrollView(
			ui.VStack(body).Alignment(ui.Center).FullWidth().Padding(ui.Padding{}.All(ui.L16)),
		).Frame(ui.Frame{Height: "calc(100dvh - 5rem)"}.FullWidth()),
	).
		BackgroundColor(ui.M1).
		Position(ui.Position{
			Type:   ui.PositionFixed,
			Left:   "0px",
			Top:    "0px",
			Right:  "0px",
			Bottom: "0px",
			ZIndex: 1000,
		}).
		Frame(ui.Frame{Width: "100dvw", Height: ui.ViewportHeight})

	return ui.Overlay(content).OnDismissRequest(closeFn)
}

// previewContent builds the actual inline preview view for the file, or the no-preview fallback.
func (c TDrive) previewContent(wnd core.Window, uc drive.UseCases, file drive.File) core.View {
	switch classifyPreview(file) {
	case previewImage:
		return ui.Image().
			URI(core.URI(drivehttp.URL(file.ID))).
			ObjectFit(ui.FitContain).
			AccessibilityLabel(file.Filename).
			Frame(ui.Frame{MaxWidth: ui.Full, Height: "calc(100dvh - 8rem)"})

	case previewVideo:
		return video.Video(core.URI(drivehttp.URL(file.ID))).
			Controls(true).
			PlaysInline(true).
			Frame(ui.Frame{MaxWidth: ui.Full, Height: "calc(100dvh - 8rem)"})

	case previewMarkdown:
		text, ok := c.readTextContent(wnd, uc, file.ID)
		if !ok {
			return c.noPreview(wnd, uc, file)
		}
		return ui.VStack(markdown.RichText(text)).
			Alignment(ui.Leading).
			Frame(ui.Frame{MaxWidth: ui.L880}.FullWidth())

	case previewText:
		text, ok := c.readTextContent(wnd, uc, file.ID)
		if !ok {
			return c.noPreview(wnd, uc, file)
		}
		return ui.CodeEditor(text).
			Language(languageForFilename(file.Filename)).
			Disabled(true).
			Frame(ui.Frame{Height: "calc(100dvh - 8rem)"}.FullWidth())

	default:
		return c.noPreview(wnd, uc, file)
	}
}

// noPreview renders the fallback shown when a file cannot be previewed, with a prominent download button.
func (c TDrive) noPreview(wnd core.Window, uc drive.UseCases, file drive.File) core.View {
	return ui.VStack(
		ui.ImageIcon(icons.ExclamationCircle),
		ui.Text(file.Filename).Font(ui.TitleMedium),
		ui.Text(StrPreviewNoPreviewDesc.Get(wnd)).Color(ui.ColorText),
		ui.PrimaryButton(func() {
			c.downloadFile(wnd, uc, file.ID)
		}).PreIcon(icons.Download).Title(StrPreviewDownload.Get(wnd)),
	).
		Gap(ui.L16).
		Alignment(ui.Center).
		Frame(ui.Frame{Height: "calc(100dvh - 10rem)"}.FullWidth())
}

// downloadFile triggers a native download of the given drive file.
func (c TDrive) downloadFile(wnd core.Window, uc drive.UseCases, fid drive.FID) {
	optFile, err := uc.Get(wnd.Subject(), fid, "")
	if err != nil {
		alert.ShowBannerError(wnd, err)
		return
	}
	if optFile.IsNone() {
		return
	}

	wnd.ExportFiles(core.ExportFilesOptions{
		ID:    string(fid),
		Files: []core.File{optFile.Unwrap()},
	})
}

// readTextContent reads up to maxTextPreviewBytes of the file and returns it as a string if it is valid utf8.
func (c TDrive) readTextContent(wnd core.Window, uc drive.UseCases, fid drive.FID) (string, bool) {
	optFile, err := uc.Get(wnd.Subject(), fid, "")
	if err != nil {
		slog.Error("preview: cannot open text file", "fid", fid, "err", err.Error())
		return "", false
	}
	if optFile.IsNone() {
		return "", false
	}

	reader, err := optFile.Unwrap().Open()
	if err != nil {
		slog.Error("preview: cannot read text file", "fid", fid, "err", err.Error())
		return "", false
	}
	defer reader.Close()

	buf, err := io.ReadAll(io.LimitReader(reader, maxTextPreviewBytes))
	if err != nil {
		slog.Error("preview: cannot read text file content", "fid", fid, "err", err.Error())
		return "", false
	}

	if !utf8.Valid(buf) {
		return "", false
	}

	return string(buf), true
}

// classifyPreview determines how a file should be previewed based on its stored mime type and file extension.
func classifyPreview(file drive.File) previewKind {
	if file.IsDir() || file.FileInfo.IsNone() {
		return previewNone
	}

	mediaType := parseMediaType(file.FileInfo.Unwrap().MimeType)
	ext := strings.ToLower(filepath.Ext(file.Filename))

	switch mediaType {
	case "image/png", "image/jpeg", "image/jpg", "image/gif", "image/webp", "image/bmp", "image/svg+xml":
		return previewImage
	case "text/markdown", "text/x-markdown":
		return previewMarkdown
	}

	if strings.HasPrefix(mediaType, "video/") {
		return previewVideo
	}

	if ext == ".md" || ext == ".markdown" {
		return previewMarkdown
	}

	if strings.HasPrefix(mediaType, "image/") {
		return previewImage
	}

	if isTextual(mediaType, ext) {
		return previewText
	}

	return previewNone
}

// parseMediaType extracts the bare media type (e.g. "text/plain") from a possibly parameterized mime string
// such as "text/plain; charset=utf-8" as produced by the `file --mime` detection.
func parseMediaType(m string) string {
	if m == "" {
		return ""
	}
	if mt, _, err := mime.ParseMediaType(m); err == nil {
		return mt
	}
	// best-effort fallback: cut at the first ';'
	if idx := strings.IndexByte(m, ';'); idx >= 0 {
		return strings.TrimSpace(strings.ToLower(m[:idx]))
	}
	return strings.TrimSpace(strings.ToLower(m))
}

// isTextual reports whether the given media type / extension denotes a textual (utf8) format that can be shown
// in the read-only code editor.
func isTextual(mediaType, ext string) bool {
	if strings.HasPrefix(mediaType, "text/") {
		return true
	}

	switch mediaType {
	case "application/json", "application/xml", "application/x-yaml", "application/yaml",
		"application/javascript", "application/x-sh", "application/x-shellscript",
		"application/toml", "application/x-toml", "image/svg+xml":
		return true
	}

	switch ext {
	case ".txt", ".md", ".markdown", ".json", ".xml", ".yaml", ".yml", ".toml", ".ini", ".conf", ".cfg",
		".csv", ".tsv", ".log", ".go", ".js", ".ts", ".jsx", ".tsx", ".css", ".scss", ".html", ".htm",
		".sh", ".bash", ".zsh", ".py", ".rb", ".java", ".kt", ".c", ".h", ".cpp", ".hpp", ".rs", ".sql",
		".env", ".properties", ".gitignore", ".dockerfile", ".adoc", ".rst", ".tex":
		return true
	}

	return false
}

// languageForFilename maps a filename extension to a code editor language identifier (best effort).
func languageForFilename(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".go":
		return "go"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".json":
		return "json"
	case ".xml", ".svg":
		return "xml"
	case ".html", ".htm":
		return "html"
	case ".css", ".scss":
		return "css"
	case ".md", ".markdown":
		return "markdown"
	case ".yaml", ".yml":
		return "yaml"
	case ".sql":
		return "sql"
	case ".sh", ".bash", ".zsh":
		return "shell"
	case ".py":
		return "python"
	default:
		return "text"
	}
}
