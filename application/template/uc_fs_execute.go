// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"testing/fstest"

	"go.wdy.de/nago/auth"
)

func NewFSExecute() FSExecute {
	return func(subject auth.Subject, fsys fs.FS, execType ExecType, options ExecOptions) (io.ReadCloser, error) {
		// TODO localize logic is missing and should be unified between fs.FS and []File

		// TODO the following is highly redundant, can we unify this?
		switch execType {
		case Unprocessed:
			if options.TemplateName != "" {
				return nil, fmt.Errorf("template is of type 'unprocessed' but a template name is given, which is not allowed")
			}

			fsys, err := cloneFS(fsys)
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
			var err error
			if execType == TreeTemplateHTML {
				tpl, err = FSParseHTML(fsys)
			} else {
				tpl, err = FSParseText(fsys)
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
			fsys, err := cloneFS(fsys)
			if err != nil {
				return nil, fmt.Errorf("cannot load template files: %w", err)
			}

			if err := applyPlainTextTemplates(fsys, options.Model); err != nil {
				return nil, fmt.Errorf("cannot apply plain text templates: %w", err)
			}

			return execTypst(fsys)
		default:
			return nil, fmt.Errorf("unknown project type: %v", execType)
		}
	}
}

func cloneFS(fset fs.FS) (fstest.MapFS, error) {
	fsys := fstest.MapFS{}
	err := fs.WalkDir(fset, ".", func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsRegular() {
			buf, err := fs.ReadFile(fset, path)
			if err != nil {
				return err
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			fsys[path] = &fstest.MapFile{
				Data:    buf,
				ModTime: info.ModTime(),
				Sys:     d,
				Mode:    info.Mode(),
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fsys, nil
}

// Apply statically applies simply and directly from the filesystem and returns it as a buffer.
func Apply(subject auth.Subject, fsys fs.FS, execType ExecType, options ExecOptions) ([]byte, error) {
	reader, err := NewFSExecute()(subject, fsys, execType, options)
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	return io.ReadAll(reader)
}
