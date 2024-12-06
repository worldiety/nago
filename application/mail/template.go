package mail

import (
	"bytes"
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application/mail/tpl"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	htmlTemplate "html/template"
	"io"
	"iter"
	textTemplate "text/template"
)

type TemplateID string

type PartType string

const (
	PartHTML PartType = "html"
	PartText PartType = "text"
)

type Template struct {
	ID          TemplateID `visible:"false"`
	Name        string
	LanguageTag string   `label:"Sprache"` // e.g. like de_DE
	Type        PartType `values:"[\"html\", \"text\"]"`
	// always go/text/template
	Subject string `label:"Betreff"`
	// either go/text/template or go/html/template to avoid injection attacks, depending on Type
	Body string `label:"Nachricht" table-visible:"false" lines:"10"`
}

func (t Template) Identity() TemplateID {
	return t.ID
}

func (t Template) WithIdentity(id TemplateID) Template {
	t.ID = id
	return t
}

func (t Template) Render(model tpl.Model) (subject string, body []byte, err error) {
	ts := textTemplate.New(fmt.Sprintf("%s:%s", t.ID, t.Name))
	if _, err := ts.Parse(t.Body); err != nil {
		return "", nil, fmt.Errorf("error parsing mail subject text template %s: %v", t.Body, err)
	}

	var tmp bytes.Buffer
	if err := ts.Execute(&tmp, model); err != nil {
		return "", nil, fmt.Errorf("error rendering mail subject template %s: %v", t.Body, err)
	}

	subject = tmp.String()

	var exec interface {
		Execute(wr io.Writer, data any) error
	}
	switch t.Type {
	case PartHTML:
		tt := textTemplate.New(fmt.Sprintf("%s:%s", t.ID, t.Name))
		if _, err := tt.Parse(t.Body); err != nil {
			return "", nil, fmt.Errorf("error parsing mail body text template %s: %v", t.Body, err)
		}

		exec = tt
	default:
		tt := htmlTemplate.New(fmt.Sprintf("%s:%s", t.ID, t.Name))
		if _, err := tt.Parse(t.Body); err != nil {
			return "", nil, fmt.Errorf("error parsing mail body html template %s: %v", t.Body, err)
		}

		exec = tt
	}

	tmp.Reset()
	if err := exec.Execute(&tmp, model); err != nil {
		return "", nil, fmt.Errorf("error executing mail body template %s: %v", t.Body, err)
	}

	return subject, tmp.Bytes(), nil
}

type TemplateRepository data.Repository[Template, TemplateID]

type FindTemplateByID func(auth.Subject, TemplateID) (std.Option[Template], error)
type DeleteTemplateByID func(auth.Subject, TemplateID) error
type FindAllTemplates func(auth.Subject) iter.Seq2[Template, error]
type SaveTemplate func(auth.Subject, Template) (TemplateID, error)
type FindTemplateByNameAndLanguage func(subject auth.Subject, name string, languageTag string) (std.Option[Template], error)

func NewFindTemplateByNameAndLanguage(repo TemplateRepository) FindTemplateByNameAndLanguage {
	return func(subject auth.Subject, name string, languageTag string) (std.Option[Template], error) {
		// TODO permission
		for template, err := range repo.All() {
			if err != nil {
				return std.None[Template](), err
			}

			if template.Name == name && template.LanguageTag == languageTag {
				return std.Some(template), nil
			}
		}

		return std.None[Template](), nil
	}
}

type RenderTemplate func(auth.Subject, Template) (TemplateID, error)

const (
	TemplateRegistered   = "nago.mail.template.registered"
	TemplateSecurityCode = "nago.mail.template.code"
)

// InitDefaultTemplates will be invoked at startup to ensure, that there are at least some reasonable default
// mail templates.
type InitDefaultTemplates func(subject auth.Subject) error

//go:embed tpl/registered.gohtml
var tplRegistered string

func NewInitDefaultTemplates(finder FindTemplateByNameAndLanguage, saveTpl SaveTemplate) InitDefaultTemplates {

	return func(subject auth.Subject) error {
		// TODO permission
		if err := upsertTpl(subject, "NAGO Kontoregistrierung", finder, saveTpl, TemplateRegistered, tplRegistered); err != nil {
			return fmt.Errorf("cannot init register template: %w", err)
		}

		if err := upsertTpl(subject, "NAGO Sicherheitscode", finder, saveTpl, TemplateSecurityCode, tplRegistered); err != nil {
			return fmt.Errorf("cannot init register template: %w", err)
		}

		return nil
	}
}

func upsertTpl(subject auth.Subject, subjectTpl string, finder FindTemplateByNameAndLanguage, saveTpl SaveTemplate, name string, tpl string) error {
	optTpl, err := finder(subject, name, "de_DE")
	if err != nil {
		return fmt.Errorf("cannot find template: %w", err)
	}

	if optTpl.IsNone() || optTpl.Unwrap().Subject == "" {
		if _, err := saveTpl(subject, Template{
			ID:          data.RandIdent[TemplateID](),
			LanguageTag: "de_DE",
			Name:        name,
			Type:        PartHTML,
			Subject:     subjectTpl,
			Body:        tpl,
		}); err != nil {
			return fmt.Errorf("cannot save template: %w", err)
		}
	}

	return nil
}
