// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package localization

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/auth"
	"golang.org/x/text/language"
)

func NewAddLanguage(res *i18n.Resources) AddLanguage {
	return func(subject auth.Subject, lang language.Tag) error {
		if err := subject.Audit(PermAddLanguage); err != nil {
			return err
		}

		res.AddLanguage(lang)
		return nil
	}
}
