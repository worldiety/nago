// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package onboarding_test

import (
	"os"
	"testing"

	"go.wdy.de/nago/application/onboarding"
)

func TestEmailFromEnv(t *testing.T) {
	const own = "NAGO_TEST_OWN_EMAIL"
	const unset = "\x00unset"

	tests := []struct {
		name      string
		env       string
		vars      map[string]string
		wantEmail string
		wantFrom  string
	}{
		{"nothing set", "", map[string]string{onboarding.EnvInitialUserEmail: unset, onboarding.EnvHubOnboardingMail: unset}, "", onboarding.EnvInitialUserEmail},
		{"both empty", "", map[string]string{onboarding.EnvInitialUserEmail: "", onboarding.EnvHubOnboardingMail: ""}, "", onboarding.EnvInitialUserEmail},
		{"only initial", "", map[string]string{onboarding.EnvInitialUserEmail: "first@example.org", onboarding.EnvHubOnboardingMail: unset}, "first@example.org", onboarding.EnvInitialUserEmail},
		{"only hub", "", map[string]string{onboarding.EnvInitialUserEmail: unset, onboarding.EnvHubOnboardingMail: "hub@example.org"}, "hub@example.org", onboarding.EnvHubOnboardingMail},
		{"initial wins over hub", "", map[string]string{onboarding.EnvInitialUserEmail: "first@example.org", onboarding.EnvHubOnboardingMail: "hub@example.org"}, "first@example.org", onboarding.EnvInitialUserEmail},
		{"invalid initial wins over hub", "", map[string]string{onboarding.EnvInitialUserEmail: "foo", onboarding.EnvHubOnboardingMail: "hub@example.org"}, "foo", onboarding.EnvInitialUserEmail},
		{"white space initial counts as empty", "", map[string]string{onboarding.EnvInitialUserEmail: "  \t", onboarding.EnvHubOnboardingMail: "hub@example.org"}, "hub@example.org", onboarding.EnvHubOnboardingMail},
		{"white space hub counts as empty", "", map[string]string{onboarding.EnvInitialUserEmail: unset, onboarding.EnvHubOnboardingMail: "  "}, "", onboarding.EnvInitialUserEmail},
		{"value as it is", "", map[string]string{onboarding.EnvInitialUserEmail: " First@Example.org ", onboarding.EnvHubOnboardingMail: unset}, " First@Example.org ", onboarding.EnvInitialUserEmail},
		{"own variable", own, map[string]string{own: "own@example.org", onboarding.EnvInitialUserEmail: "first@example.org", onboarding.EnvHubOnboardingMail: "hub@example.org"}, "own@example.org", own},
		{"own variable only, empty", own, map[string]string{own: "", onboarding.EnvInitialUserEmail: "first@example.org", onboarding.EnvHubOnboardingMail: "hub@example.org"}, "", own},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.vars {
				t.Setenv(k, "")
				if v == unset {
					if err := os.Unsetenv(k); err != nil {
						t.Fatal(err)
					}
					continue
				}
				t.Setenv(k, v)
			}

			email, from := onboarding.EmailFromEnv(tt.env)
			if email != tt.wantEmail || from != tt.wantFrom {
				t.Fatalf("EmailFromEnv(%q) = %q, %q; want %q, %q", tt.env, email, from, tt.wantEmail, tt.wantFrom)
			}
		})
	}
}
