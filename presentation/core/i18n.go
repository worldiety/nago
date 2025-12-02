// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

// DefaultStr asserts an i18n english and german string and panics if it has already been declared.
func DefaultStr(key i18n.Key, en, de string) i18n.StrHnd {
	return i18n.MustString(key, i18n.Values{language.English: en, language.German: de})
}
