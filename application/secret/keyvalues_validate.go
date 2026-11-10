// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"go.wdy.de/nago/pkg/std"
	"golang.org/x/net/http/httpguts"
)

// Validatable is an optional extension of [Credentials]. [CreateSecret] and [UpdateMyCredentials] store the result
// of Validated instead of the given credentials, after the stored write-only values have been merged. The text of
// an error must never contain a value.
type Validatable interface {
	Validated() (Credentials, error)
}

func validated(credentials Credentials) (Credentials, error) {
	if v, ok := credentials.(Validatable); ok {
		return v.Validated()
	}

	return credentials, nil
}

// MaxLogoSVGBytes is the limit of [KeyValues.LogoSVG].
const MaxLogoSVGBytes = 64 << 10

var keyValueKeyRegex = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]*$`)

func invalidKeyValues(format string, args ...any) error {
	return std.NewLocalizedError("Secret nicht gespeichert", fmt.Sprintf(format, args...))
}

func invalidLogo(message string) error {
	return std.NewLocalizedError("Logo nicht übernommen", message)
}

// Validated implements [Validatable]. It fills missing identifiers and refuses what a connection could misuse.
func (kv KeyValues) Validated() (Credentials, error) {
	kv = kv.WithIDs()
	if strings.TrimSpace(kv.Name) == "" {
		return nil, invalidKeyValues("Der Name darf nicht leer sein.")
	}

	if _, err := ParseLogoSVG([]byte(kv.LogoSVG)); kv.LogoSVG != "" && err != nil {
		return nil, err
	}

	sectionIDs := map[string]bool{}
	fieldIDs := map[string]bool{}
	keys := map[string]bool{}
	headers := map[string]string{} // by name, the key which sets it
	for _, section := range kv.Sections {
		if sectionIDs[section.ID] {
			return nil, invalidKeyValues("Die Kennung eines Abschnitts ist nicht eindeutig.")
		}

		sectionIDs[section.ID] = true
		for _, field := range section.Fields {
			if fieldIDs[field.ID] {
				return nil, invalidKeyValues("Die Kennung des Feldes '%s' ist nicht eindeutig.", field.Key)
			}

			fieldIDs[field.ID] = true
			if err := field.validate(); err != nil {
				return nil, err
			}

			if keys[field.Key] {
				return nil, invalidKeyValues("Der Schlüssel '%s' ist mehrfach vergeben.", field.Key)
			}

			keys[field.Key] = true
			if header := headerSetBy(field.Key); header != "" {
				if other, taken := headers[header]; taken && !(isBasicAuthKey(other) && isBasicAuthKey(field.Key)) {
					return nil, invalidKeyValues("Die Schlüssel '%s' und '%s' setzen denselben Header '%s'.", other, field.Key, header)
				}

				headers[header] = field.Key
			}
		}
	}

	return kv, nil
}

// headerSetBy returns the header which the key sets, because only one value of a header can reach the server.
func headerSetBy(key string) string {
	if key == KeyHeaderBearerAuth || isBasicAuthKey(key) {
		return "Authorization"
	}

	header, _ := HeaderForKey(key)
	return header
}

func isBasicAuthKey(key string) bool {
	return key == KeyHeaderBasicAuthUser || key == KeyHeaderBasicAuthPassword
}

func (f KeyValueField) validate() error {
	if f.Key == "" {
		return invalidKeyValues("Jedes Feld benötigt einen Schlüssel.")
	}

	if !keyValueKeyRegex.MatchString(f.Key) {
		return invalidKeyValues("Der Schlüssel '%s' ist ungültig. Er muss mit einem Buchstaben beginnen und darf danach nur Buchstaben, Ziffern und die Zeichen _ . - enthalten.", f.Key)
	}

	switch f.Kind {
	case FieldString, FieldMultiLine, FieldInt, FieldBool, FieldURL:
	default:
		return invalidKeyValues("Das Feld '%s' hat eine unbekannte Art.", f.Key)
	}

	switch f.Protection {
	case Visible, Concealed:
	case WriteOnly:
		if !WriteOnlyAllowed(f.Kind) {
			return invalidKeyValues("Das Feld '%s' kann den Schutz 'nur schreibbar' nicht erhalten, das ist nur für Text und mehrzeiligen Text möglich.", f.Key)
		}
	default:
		return invalidKeyValues("Das Feld '%s' hat einen unbekannten Schutz.", f.Key)
	}

	if err := f.validateFixedKey(); err != nil {
		return err
	}

	if !ValueFitsKind(f.Kind, f.Value) {
		return invalidKeyValues("Der Wert des Feldes '%s' passt nicht zu seiner Art.", f.Key)
	}

	return nil
}

func (f KeyValueField) validateFixedKey() error {
	if f.Key == KeyBaseURL {
		if f.Kind != FieldURL || f.Protection == WriteOnly {
			return invalidKeyValues("Das Feld '%s' muss von der Art URL sein und darf den Schutz 'nur schreibbar' nicht haben.", f.Key)
		}

		if f.Value != "" && !validBaseURL(f.Value) {
			return invalidKeyValues("Der Wert des Feldes '%s' muss eine http- oder https-Adresse mit Host und ohne Zugangsdaten sein.", f.Key)
		}

		return nil
	}

	if strings.HasPrefix(f.Key, KeyPrefixHeader) {
		isAuth := f.Key == KeyHeaderBearerAuth || isBasicAuthKey(f.Key)
		if _, ok := HeaderForKey(f.Key); !isAuth && !ok {
			return invalidKeyValues("Der Schlüssel '%s' benennt keinen Header, den eine Verbindung setzen darf.", f.Key)
		}

		if !httpguts.ValidHeaderFieldValue(f.Value) {
			return invalidKeyValues("Der Wert des Feldes '%s' muss einzeilig sein und darf keine Steuerzeichen enthalten.", f.Key)
		}

		return nil
	}

	if strings.HasPrefix(f.Key, KeyPrefixQuery) {
		if _, ok := QueryForKey(f.Key); !ok {
			return invalidKeyValues("Der Schlüssel '%s' benennt keinen Parameter.", f.Key)
		}

		if strings.ContainsAny(f.Value, "\r\n") {
			return invalidKeyValues("Der Wert des Feldes '%s' muss einzeilig sein.", f.Key)
		}
	}

	return nil
}

// WriteOnlyAllowed returns true, if a field of the kind may be [WriteOnly].
func WriteOnlyAllowed(kind FieldKind) bool {
	return kind == FieldString || kind == FieldMultiLine
}

// ValueFitsKind returns true, if the value is empty or the canonical text of the kind.
func ValueFitsKind(kind FieldKind, value string) bool {
	if value == "" {
		return true
	}

	switch kind {
	case FieldString:
		return !strings.ContainsAny(value, "\r\n")
	case FieldMultiLine:
		return true
	case FieldInt:
		i, err := strconv.ParseInt(value, 10, 64)
		return err == nil && strconv.FormatInt(i, 10) == value
	case FieldBool:
		return value == "true" || value == "false"
	case FieldURL:
		u, err := url.Parse(value)
		return err == nil && u.Scheme != "" && (u.Host != "" || u.Opaque != "")
	}

	return false
}

func validBaseURL(value string) bool {
	u, err := url.Parse(value)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != "" && u.User == nil
}

// ParseLogoSVG checks a file for [KeyValues.LogoSVG] and returns the document from its root element on. What is in
// front of it, like an XML declaration or a long comment, is dropped, because an image of package ui recognizes an SVG
// only by the beginning of its data. There the frontend also removes everything active from it.
func ParseLogoSVG(raw []byte) (string, error) {
	if len(raw) > MaxLogoSVGBytes {
		return "", invalidLogo(fmt.Sprintf("Das Logo darf höchstens %d KiB groß sein.", MaxLogoSVGBytes>>10))
	}

	doc := strings.TrimPrefix(string(raw), "\xef\xbb\xbf")
	offset, ok := svgRootOffset(doc)
	if !ok {
		return "", invalidLogo("Das Logo muss ein SVG-Dokument sein.")
	}

	return doc[offset:], nil
}

// svgRootOffset returns the position of the root element of the document, if that is an svg element.
func svgRootOffset(doc string) (int64, bool) {
	dec := xml.NewDecoder(strings.NewReader(doc))
	dec.Strict = false // e.g. the entities in the root element of an Illustrator export
	for {
		offset := dec.InputOffset()
		token, err := dec.RawToken()
		if err != nil {
			return 0, false
		}

		switch t := token.(type) {
		case xml.StartElement:
			return offset, t.Name.Local == "svg"
		case xml.CharData:
			if strings.TrimSpace(string(t)) != "" {
				return 0, false
			}
		}
	}
}
