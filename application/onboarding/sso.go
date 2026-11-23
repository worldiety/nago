// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package onboarding

import (
	"regexp"
	"slices"
	"strings"

	"go.wdy.de/nago/application/user"
)

// DefaultNLSServer is the Nago login service, which the setup offers for the single sign-on.
const DefaultNLSServer = "https://login.worldiety.nago.app"

// SSOAccessKind tells who may sign in by single sign-on after the setup.
type SSOAccessKind int

const (
	// SSOOnlyAdmin allows the mail address of the first user only.
	SSOOnlyAdmin SSOAccessKind = iota
	// SSODomain allows every address of the domain of the first user, see [CompanyDomain].
	SSODomain
	// SSOSelected allows the first user and the addresses and domains of [SSOAccess.Entries].
	SSOSelected
	// SSOEveryone allows every account of the login service.
	SSOEveryone
)

// SSOAccess describes who may sign in by single sign-on after the setup. A person, who signs in for the first time,
// gets an account without roles.
type SSOAccess struct {
	Kind    SSOAccessKind
	Entries []AccessEntry // only for SSOSelected
}

// AccessEntry is a mail address or a domain, whose addresses may sign in.
type AccessEntry struct {
	Email  user.Email // set for a single address
	Domain string     // set for all addresses of a domain
}

func (e AccessEntry) String() string {
	if e.Email != "" {
		return string(e.Email)
	}

	return "@" + e.Domain
}

func (e AccessEntry) pattern() string {
	if e.Email != "" {
		return emailPattern(e.Email)
	}

	return domainPattern(e.Domain)
}

// InvalidEntry is a part of the text of [ParseAccessEntries], which looks like a mail address or a domain, but is
// none or must not be used.
type InvalidEntry struct {
	Text string
	// PublicDomain tells that the text is the domain of a public mail provider, which would allow anybody.
	PublicDomain bool
}

// ParseAccessEntries reads mail addresses and domains from text, as people type or paste them: separated by commas,
// semicolons, spaces or lines, also as "Name <mail>", "mailto:mail", "@domain" or "www.domain". Words without an @
// and without a dot, like names, are skipped. It returns the entries without duplicates in their order and the parts,
// which look like an address or a domain but are none, or which are the domain of a public mail provider.
func ParseAccessEntries(text string) (entries []AccessEntry, invalid []InvalidEntry) {
	tokens := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		switch r {
		case ',', ';', '<', '>', '(', ')', '[', ']', '"', '\'', '`':
			return true
		}

		return r <= ' ' || r == ' '
	})

	for _, token := range tokens {
		token = strings.TrimPrefix(token, "mailto:")
		token = strings.Trim(token, ".:")
		if token == "" || !strings.ContainsAny(token, "@.") {
			continue
		}

		var entry AccessEntry
		switch local, domain, found := strings.Cut(token, "@"); {
		case found && local != "":
			if email := user.Email(token); email.Valid() && validDomain(domain) {
				entry.Email = email
			}
		case found:
			if validDomain(domain) {
				entry.Domain = domain
			}
		default:
			domain = strings.TrimPrefix(strings.TrimPrefix(token, "https://"), "http://")
			domain, _, _ = strings.Cut(domain, "/")
			domain = strings.TrimPrefix(domain, "www.")
			if validDomain(domain) {
				entry.Domain = domain
			}
		}

		switch {
		case entry == AccessEntry{}:
			invalid = append(invalid, InvalidEntry{Text: token})
		case entry.Domain != "" && publicMailDomain(entry.Domain):
			invalid = append(invalid, InvalidEntry{Text: token, PublicDomain: true})
		case !slices.Contains(entries, entry):
			entries = append(entries, entry)
		}
	}

	return entries, invalid
}

// AllowList returns the patterns of [user.Settings.SSOAllowList], which grant the access. The first user is always
// allowed. An empty list allows everybody.
func (a SSOAccess) AllowList(admin user.Email) []string {
	switch a.Kind {
	case SSOEveryone:
		return nil
	case SSODomain:
		if domain, ok := CompanyDomain(admin); ok {
			return []string{domainPattern(domain)}
		}
	case SSOSelected:
		list := []string{emailPattern(admin)}
		for _, entry := range a.Entries {
			if p := entry.pattern(); !slices.Contains(list, p) {
				list = append(list, p)
			}
		}

		return list
	}

	return []string{emailPattern(admin)}
}

// CompanyDomain returns the domain of the address, unless it is the domain of a public mail provider, which must not
// be allowed as a whole.
func CompanyDomain(email user.Email) (string, bool) {
	_, domain, ok := strings.Cut(strings.ToLower(string(email)), "@")
	if !ok || !validDomain(domain) || publicMailDomain(domain) {
		return "", false
	}

	return domain, true
}

func emailPattern(email user.Email) string {
	return "^" + regexp.QuoteMeta(strings.ToLower(string(email))) + "$"
}

func domainPattern(domain string) string {
	return "^[^@]+@" + regexp.QuoteMeta(domain) + "$"
}

var domainRegex = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,}$`)

func validDomain(domain string) bool {
	return len(domain) <= 253 && domainRegex.MatchString(domain)
}

// publicMailDomains are providers, whose addresses anybody can get.
var publicMailDomains = []string{
	"gmail.com", "googlemail.com", "outlook.com", "outlook.de", "hotmail.com", "hotmail.de", "live.com", "live.de",
	"msn.com", "yahoo.com", "yahoo.de", "ymail.com", "aol.com", "aol.de", "icloud.com", "me.com", "mac.com",
	"gmx.de", "gmx.net", "gmx.at", "gmx.ch", "gmx.com", "web.de", "t-online.de", "freenet.de", "posteo.de",
	"posteo.net", "mailbox.org", "proton.me", "protonmail.com", "pm.me", "tutanota.com", "tutanota.de", "tuta.io",
	"arcor.de", "online.de", "email.de", "mail.de", "mail.com", "yandex.com", "yandex.ru", "zoho.com", "gmx.eu",
}

func publicMailDomain(domain string) bool {
	return slices.Contains(publicMailDomains, domain)
}
