package mail

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewInitDefaultTemplates(finder FindTemplateByNameAndLanguage, saveTpl SaveTemplate) InitDefaultTemplates {

	return func(subject auth.Subject) error {
		if err := subject.Audit(PermInitDefaultTemplates); err != nil {
			return err
		}
		
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
