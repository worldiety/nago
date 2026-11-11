// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tplmail

import (
	"html/template"
	"strings"
	"testing"
)

func TestRegistrationTemplates(t *testing.T) {
	tpl, err := template.ParseFS(Files, "*.gohtml")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		model any
		want  string
	}{
		{string(RegistrationLink), RegistrationLinkModel{RegisterURL: "https://example.com/account/register?id=a&code=b", ValidHours: 24, ApplicationName: "Demo"}, "https://example.com/account/register?id=a&amp;code=b"},
		{string(RegistrationLinkSubject), RegistrationLinkModel{ApplicationName: "Demo"}, "Konto erstellen bei Demo"},
		{string(AlreadyRegistered), AlreadyRegisteredModel{LoginURL: "https://example.com/account/login", ApplicationName: "Demo"}, "https://example.com/account/login"},
		{string(AlreadyRegisteredSubject), AlreadyRegisteredModel{ApplicationName: "Demo"}, "Dein Konto bei Demo"},
	}

	for _, c := range cases {
		var sb strings.Builder
		if err := tpl.ExecuteTemplate(&sb, c.name, c.model); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}

		if !strings.Contains(sb.String(), c.want) {
			t.Fatalf("%s: expected %q in %q", c.name, c.want, sb.String())
		}
	}
}
