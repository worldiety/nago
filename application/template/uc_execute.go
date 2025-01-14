package template

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	htmlTemplate "html/template"
	"io"
	"io/fs"
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

		case TreeTemplate:
			var tpl templater
			if project.HasHTML() {
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

			buf, err := zipIt(fsys)
			if err != nil {
				return nil, fmt.Errorf("cannot zip templates: %w", err)
			}

			return io.NopCloser(bytes.NewReader(buf)), nil
		default:
			return nil, fmt.Errorf("unknown project type: %v", project.Type)
		}
	}
}

func execTypst(fsys fs.FS) (io.ReadCloser, error) {
	panic("not implemented")
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
