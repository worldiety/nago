// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package settings

import (
	"reflect"
	"testing"

	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var testTitle = i18n.MustString("nago.settings.test.title", i18n.Values{
	language.German:  "Testtitel",
	language.English: "Test title",
})

type testBundler struct{ b *i18n.Bundle }

func (t testBundler) Bundle() *i18n.Bundle { return t.b }

type keyedSettings struct {
	_ any `title:"nago.settings.test.title" description:"Ein wörtlicher Text"`
}

func TestMetaDataLocalize(t *testing.T) {
	_ = testTitle
	b, ok := i18n.Default.MatchBundle(language.German)
	if !ok {
		t.Fatal("no german bundle")
	}

	meta := ReadMetaData(reflect.TypeFor[keyedSettings]()).Localize(testBundler{b})
	if meta.Title != "Testtitel" {
		t.Errorf("title not localized: %q", meta.Title)
	}

	if meta.Description != "Ein wörtlicher Text" {
		t.Errorf("literal description changed: %q", meta.Description)
	}

	if got := ReadMetaData(reflect.TypeFor[keyedSettings]()).Localize(nil); got.Title != "nago.settings.test.title" {
		t.Errorf("nil bundler must not change anything: %q", got.Title)
	}
}
