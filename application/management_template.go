// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"go.wdy.de/nago/application/template"
	uitemplate "go.wdy.de/nago/application/template/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"io"
)

// TemplateManagement is a nago system(Template Management).
// Template Management is used for creating, editing, and managing reusable templates.
// It provides a centralized way to separate content from code and supports multiple output formats,
// including Go HTML templates for emails, plain text, and various text-to-PDF workflows (Typst, LaTeX, AsciiDoc).
// The system is primarily used by other modules such as Mail Management for standardized emails,
// but can also support document generation (e.g., reports, certificates, invoices).
// By centralizing template logic, it increases flexibility, maintainability, and consistency across the platform.
type TemplateManagement struct {
	UseCases template.UseCases
	Pages    uitemplate.Pages
}

func (c *Configurator) TemplateManagement() (TemplateManagement, error) {
	if c.templateManagement == nil {
		// note: we intentionally use a file store here, e.g. due to large images or other huge attachments
		// which may be really arbitrary
		fileStore, err := c.FileStore("nago.template.file")
		if err != nil {
			return TemplateManagement{}, fmt.Errorf("cannot get file store: %w", err)
		}

		projectStore, err := c.EntityStore("nago.template.project")
		if err != nil {
			return TemplateManagement{}, fmt.Errorf("cannot get entity store: %w", err)
		}

		projectRepo := json.NewSloppyJSONRepository[template.Project, template.ID](projectStore)

		uc := template.NewUseCases(fileStore, projectRepo)
		c.templateManagement = &TemplateManagement{
			UseCases: uc,
			Pages: uitemplate.Pages{
				Projects:   "admin/template/projects",
				NewProject: "admin/template/new",
				Editor:     "admin/template/edit",
			},
		}

		c.RootViewWithDecoration(c.templateManagement.Pages.Projects, func(wnd core.Window) core.View {
			return uitemplate.ProjectPickerPage(wnd, c.templateManagement.Pages, uc.FindAll, uc.Delete)
		})

		c.RootViewWithDecoration(c.templateManagement.Pages.NewProject, func(wnd core.Window) core.View {
			return uitemplate.NewProjectPage(wnd, c.templateManagement.Pages, uc.Create)
		})

		c.RootViewWithDecoration(c.templateManagement.Pages.Editor, func(wnd core.Window) core.View {
			return uitemplate.PageEditor(wnd, uc)
		})
	}

	return *c.templateManagement, nil
}

// TemplateString renders a tree template.
func (c *Configurator) TemplateString(subject auth.Subject, id template.ID, name template.DefinedTemplateName, model any) (string, error) {
	tpls, err := c.TemplateManagement()
	if err != nil {
		return "", err
	}

	r, err := tpls.UseCases.Execute(subject, id, template.ExecOptions{
		Context:      c.Context(),
		TemplateName: name,
		Model:        model,
	})

	if err != nil {
		return "", fmt.Errorf("cannot execute mail template: %w", err)
	}

	defer r.Close()
	buf, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("cannot read mail template: %w", err)
	}

	return string(buf), nil
}
