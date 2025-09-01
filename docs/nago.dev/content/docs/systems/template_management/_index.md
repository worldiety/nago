---
title: Template Management
galleryOverview:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/template_management/galleries/overview.png"
galleryMailTemplates:
  - src: "/images/systems/template_management/galleries/email_templates/projects.png"
  - src: "/images/systems/template_management/galleries/email_templates/edit.png"
galleryPDFTemplates:
  - src: "/images/systems/template_management/galleries/pdf_templates/overview.png"
  - src: "/images/systems/template_management/galleries/pdf_templates/create.png"
  - src: "/images/systems/template_management/galleries/pdf_templates/create_2.png"
---
The Template Management system provides a centralized way to create, edit, and manage reusable templates.  
It separates content from code and ensures consistent formatting and styling across the platform.  
Templates are used by other systems such as [Mail Management](../mail_management/) for emails or document workflows for generating reports, invoices, or certificates.

{{< swiper name="galleryOverview" loop="false" >}}

## Functional areas
Template Management offers the following key functions:

### Email templates
- Default templates for common user workflows such as registration, account verification, and password reset
- Written in Go HTML (`.gohtml`) format with placeholders for dynamic data
- Support for full HTML and CSS styling
- Templates are globally available to all authorized users

{{< swiper name="galleryMailTemplates" loop="false" >}}

### PDF and document templates
Template Management supports different project types for document generation.  
These allow integration into workflows beyond email, such as automated reporting or certificate creation:

- **Generic** – a project without built-in template execution, for custom processing
- **Text to Text** – plain text templates
- **HTML to HTML** – HTML-based templates
- **Typst to PDF** – Typst projects rendered into PDF
- **LaTeX to PDF** – LaTeX projects rendered into PDF
- **AsciiDoc to PDF** – AsciiDoc projects rendered into PDF

{{< swiper name="galleryPDFTemplates" loop="false" >}}

### File management
- Each template project creates its own file system
- Files can be created, imported, and exported
- Shared partials (e.g., headers, footers, styles) can be reused across templates
- Especially useful for email templates where header and footer layouts are standardized

## Dependencies
**Requires:**
- None

**Is required by:**
- [Mail Management](../mail_management/)

## Activation
This system is activated via:
```go
std.Must(cfg.TemplateManagement())
```

```go
templateManagement := std.Must(cfg.TemplateManagement())
```