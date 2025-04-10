// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	htmlTemplate "html/template"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing/fstest"
	textTemplate "text/template"
	"time"
)

func NewExecute(files blob.Store, repository Repository) Execute {
	return func(subject auth.Subject, id ID, options ExecOptions) (io.ReadCloser, error) {
		if err := subject.Audit(PermExecute); err != nil {
			return nil, err
		}

		if options.Context == nil {
			options.Context = context.Background()
		}

		optProject, err := repository.FindByID(id)
		if err != nil {
			return nil, fmt.Errorf("cannot load project: %w", err)
		}

		if optProject.IsNone() {
			return nil, fmt.Errorf("cannot load project: project is gone")
		}

		project := optProject.Unwrap()
		fileSet := project.Localize(options.Language)

		switch project.Type {
		case Unprocessed:
			if options.TemplateName != "" {
				return nil, fmt.Errorf("template is of type 'unprocessed' but a template name is given, which is not allowed")
			}

			fsys, err := loadFS(files, fileSet)
			if err != nil {
				return nil, fmt.Errorf("cannot load template files: %w", err)
			}

			buf, err := zipIt(fsys)
			if err != nil {
				return nil, fmt.Errorf("cannot zip template files: %w", err)
			}

			return io.NopCloser(bytes.NewReader(buf)), nil

		case TreeTemplatePlain, TreeTemplateHTML:
			var tpl templater
			if project.Type == TreeTemplateHTML {
				tpl, err = ParseHTML(files, fileSet)
			} else {
				tpl, err = ParseText(files, fileSet)
			}

			if err != nil {
				return nil, fmt.Errorf("cannot parse tree template files: %w", err)
			}

			var buf bytes.Buffer
			if options.TemplateName == "" {
				if err := tpl.Execute(&buf, options.Model); err != nil {
					return nil, fmt.Errorf("cannot execute anon tree template %w", err)
				}
			} else {
				if err := tpl.ExecuteTemplate(&buf, options.TemplateName, options.Model); err != nil {
					return nil, fmt.Errorf("cannot execute '%s' tree template: %w", options.TemplateName, err)
				}
			}

			return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
		case AsciidocPDF, LatexPDF, TypstPDF:
			fsys, err := loadFS(files, fileSet)
			if err != nil {
				return nil, fmt.Errorf("cannot load template files: %w", err)
			}

			if err := applyPlainTextTemplates(fsys, options.Model); err != nil {
				return nil, fmt.Errorf("cannot apply plain text templates: %w", err)
			}

			return execTypst(fsys)
			/*buf, err := zipIt(fsys)
			if err != nil {
				return nil, fmt.Errorf("cannot zip templates: %w", err)
			}

			return io.NopCloser(bytes.NewReader(buf)), nil*/
		default:
			return nil, fmt.Errorf("unknown project type: %v", project.Type)
		}
	}
}

func execTypst(fsys fs.FS) (io.ReadCloser, error) {
	mainFile, err := bestMainTypstCandiate(fsys)
	if err != nil {
		return nil, err
	}

	tmpDir := filepath.Join(os.TempDir(), "typst", data.RandIdent[string]())
	_ = os.MkdirAll(tmpDir, 0700) // security note: do not allow that others read our directory
	defer os.RemoveAll(tmpDir)

	typstExec, ok := which("typst")
	if !ok {
		// TODO try remote rendering through wdy render host
		return nil, fmt.Errorf("cannot find typst executable")
	}

	if err := copyFS(fsys, tmpDir); err != nil {
		return nil, fmt.Errorf("cannot copy typst files: %w", err)
	}

	cmd := exec.Command(typstExec, "compile", mainFile)
	cmd.Dir = tmpDir
	buf, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("failed to execute typst command", "cmd", cmd, "buf", string(buf))
		return nil, fmt.Errorf("cannot execute typst command: %w:\n%s", err, string(buf))
	}

	pdfName := filepath.Join(tmpDir, mainFile[:len(mainFile)-4]+".pdf")
	return os.Open(pdfName)
}

func bestMainTypstCandiate(fsys fs.FS) (string, error) {
	var candidates []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(strings.ToLower(d.Name()), ".typ") {
			candidates = append(candidates, path)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	slices.Sort(candidates)
	if len(candidates) == 0 {
		return "", fmt.Errorf("cannot find any typst file (*.typ)")
	}

	return candidates[0], nil
}

func which(what string) (string, bool) {
	var staticLookups = []string{"/bin/typst", "/opt/homebrew/bin/typst"}
	for _, path := range staticLookups {
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}

	cmd := exec.Command("which", what)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("cannot find which executable in $PATH")
		return "", false
	}

	if len(buf) == 0 {
		return "", false
	}

	return strings.TrimSpace(string(buf)), true
}

func loadFS(store blob.Store, fset []File) (fstest.MapFS, error) {
	fsys := fstest.MapFS{}
	for _, file := range fset {
		optBuf, err := blob.Get(store, file.Blob)
		if err != nil {
			return nil, fmt.Errorf("cannot load blob '%s': %v", file.Blob, err)
		}

		if optBuf.IsNone() {
			return nil, fmt.Errorf("cannot load blob '%s': blob is gone", file.Blob)
		}

		fsys[file.Filename] = &fstest.MapFile{
			Data:    optBuf.Unwrap(),
			ModTime: file.LastMod,
			Sys:     file,
			Mode:    os.ModePerm,
		}
	}

	return fsys, nil
}

func zipIt(fsys fs.FS) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	if err := writer.AddFS(fsys); err != nil {
		return nil, fmt.Errorf("failed to create multi-file project zip: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

type templater interface {
	Execute(wr io.Writer, data any) error
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

// ParseHTML loads all those Files which look like a template.
func ParseHTML(files blob.Store, fset []File) (*htmlTemplate.Template, error) {
	root := htmlTemplate.New("")
	var tpl *htmlTemplate.Template
	for _, file := range fset {
		if !file.IsTemplate() {
			continue
		}

		optBuf, err := blob.Get(files, file.Blob)
		if err != nil {
			return nil, fmt.Errorf("cannot read file '%s': %w", file.Filename, err)
		}

		if optBuf.IsNone() {
			return nil, fmt.Errorf("file '%s' is gone", file.Filename)
		}

		tpl, err = root.New(file.Filename).Parse(string(optBuf.Unwrap()))

		if err != nil {
			return nil, fmt.Errorf("cannot parse html file '%s': %w", file.Filename, err)
		}
	}

	if tpl == nil {
		return nil, fmt.Errorf("no template found")
	}

	return root, nil
}

func applyPlainTextTemplates(fsys fstest.MapFS, model any) error {
	for filename, file := range fsys {
		if !(strings.HasSuffix(filename, ".gohtml") || strings.HasSuffix(filename, ".tpl")) {
			continue
		}

		tpl, err := textTemplate.New(filename).Parse(string(file.Data))
		if err != nil {
			return fmt.Errorf("cannot parse file '%s': %w", filename, err)
		}

		var buf bytes.Buffer
		if err := tpl.Execute(&buf, model); err != nil {
			return fmt.Errorf("cannot execute template '%s': %w", filename, err)
		}

		fsys[cleanName(filename)] = &fstest.MapFile{
			Data:    buf.Bytes(),
			ModTime: time.Now(),
			Mode:    os.ModePerm,
		}
	}

	return nil
}

// ParseText loads all those Files which look like a template.
func ParseText(files blob.Store, fset []File) (*textTemplate.Template, error) {
	root := textTemplate.New("")
	var tpl *textTemplate.Template
	for _, file := range fset {
		if !file.IsTemplate() {
			continue
		}

		optBuf, err := blob.Get(files, file.Blob)
		if err != nil {
			return nil, fmt.Errorf("cannot read file '%s': %w", file.Filename, err)
		}

		if optBuf.IsNone() {
			return nil, fmt.Errorf("file '%s' is gone", file.Filename)
		}

		tpl, err = root.New(file.Filename).Parse(string(optBuf.Unwrap()))

		if err != nil {
			return nil, fmt.Errorf("cannot parse html file '%s': %w", file.Filename, err)
		}
	}

	if tpl == nil {
		return nil, fmt.Errorf("no template found")
	}

	return root, nil
}

func copyFS(srcFS fs.FS, destDir string) error {
	// Alle Dateien und Verzeichnisse im fs.FS durchlaufen
	return fs.WalkDir(srcFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, path)

		if d.IsDir() {
			// Falls es ein Verzeichnis ist, erstellen
			return os.MkdirAll(destPath, 0755)
		}

		// Datei öffnen
		srcFile, err := srcFS.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Datei im lokalen Verzeichnis erstellen
		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		// Dateiinhalt kopieren
		_, err = io.Copy(destFile, srcFile)
		return err
	})
}
