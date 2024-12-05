package mail

import (
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application/mail/tpl"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"iter"
)

type TemplateID string

type PartType string

const (
	PartHTML PartType = "html"
	PartText PartType = "text"
)

type Template struct {
	ID          TemplateID
	LanguageTag string // e.g. like de_DE
	Name        string
	Type        PartType
	Subject     string // a go/text/template
	Template    string // either go/text/template or go/html/template to avoid injection attacks
}

func (t Template) Identity() TemplateID {
	return t.ID
}

func (t Template) WithIdentity(id TemplateID) Template {
	t.ID = id
	return t
}

func (t Template) Render(model tpl.Model) []byte {
	/*switch t.Type {
	case PartHTML:
		tpl := textTemplate.New(fmt.Sprintf("%s:%s", t.ID, t.Name))
		tpl.Parse()
	default:
		tpl := textTemplate.New(fmt.Sprintf("%s:%s", t.ID, t.Name))
	}*/
	panic("implement me")
}

type TemplateRepository data.Repository[Template, TemplateID]

type FindTemplateByID func(auth.Subject, ID) (std.Option[Template], error)
type DeleteTemplateByID func(auth.Subject, ID) error
type FindAllTemplates func(auth.Subject) iter.Seq2[Template, error]
type SaveTemplate func(auth.Subject, Template) (ID, error)

type RenderTemplate func(auth.Subject, Template) (ID, error)

const (
	TemplateRegistered   TemplateID = "nago.mail.template.registered"
	TemplateSecurityCode TemplateID = "nago.mail.template.code"
)

type InitDefaultTemplates func(subject auth.Subject) error

//go:embed tpl/registered.gohtml
var tplRegistered string

func NewInitDefaultTemplates(repo TemplateRepository, saveTpl SaveTemplate) InitDefaultTemplates {

	return func(subject auth.Subject) error {
		// TODO permission
		if err := upsertTpl(subject, repo, saveTpl, TemplateRegistered, "Konto registriert", tplRegistered); err != nil {
			return fmt.Errorf("cannot init register template: %w", err)
		}

		if err := upsertTpl(subject, repo, saveTpl, TemplateSecurityCode, "Konto registriert", tplRegistered); err != nil {
			return fmt.Errorf("cannot init register template: %w", err)
		}

		return nil
	}
}

func upsertTpl(subject auth.Subject, repo TemplateRepository, saveTpl SaveTemplate, id TemplateID, name string, tpl string) error {
	optTpl, err := repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("cannot find template: %w", err)
	}

	if optTpl.IsNone() || optTpl.Unwrap().Template == "" {
		if _, err := saveTpl(subject, Template{
			ID:          id,
			LanguageTag: "de_DE",
			Name:        name,
			Type:        PartHTML,
			Template:    tpl,
		}); err != nil {
			return fmt.Errorf("cannot save template: %w", err)
		}
	}

	return nil
}
