// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"iter"
	"net/http"
	"slices"
	"strings"

	"go.wdy.de/nago/pkg/data"
	"golang.org/x/net/http/httpguts"
)

// The keys of a [KeyValues], which a connection consumes, see package gostrictnet. A program never sees them, all
// other keys which are not write-only are readable by programs.
const (
	// KeyBaseURL is the address of the connection. If it changes, the stored write-only values are gone.
	KeyBaseURL = "BASE_URL"
	// KeyHeaderBearerAuth is the token of an Authorization Bearer header.
	KeyHeaderBearerAuth = "HEADER_BEARER_AUTH"
	// KeyHeaderBasicAuthUser is the user name of an Authorization Basic header.
	KeyHeaderBasicAuthUser = "HEADER_BASIC_AUTH_USER"
	// KeyHeaderBasicAuthPassword is the password of an Authorization Basic header.
	KeyHeaderBasicAuthPassword = "HEADER_BASIC_AUTH_PASSWORD"
	// KeyPrefixHeader starts the key of a header. An underscore of the rest is a hyphen, so that HEADER_X_API_KEY
	// sets the header X-Api-Key.
	KeyPrefixHeader = "HEADER_"
	// KeyPrefixQuery starts the key of a query parameter, whose name is the rest verbatim.
	KeyPrefixQuery = "QUERY_"
)

// FieldKind is the type of the value of a [KeyValueField].
type FieldKind string

const (
	FieldString    FieldKind = "string"
	FieldMultiLine FieldKind = "multiline"
	FieldInt       FieldKind = "int"
	FieldBool      FieldKind = "bool"
	FieldURL       FieldKind = "url"
)

// FieldProtection tells who may read the value of a [KeyValueField].
type FieldProtection string

const (
	// Visible values are shown and readable by programs.
	Visible FieldProtection = ""
	// Concealed values are masked in the user interface, but stay readable by programs.
	Concealed FieldProtection = "concealed"
	// WriteOnly values never leave the server again, see [WithoutWriteOnly]. A value can only be replaced and only
	// a connection may use it.
	WriteOnly FieldProtection = "writeonly"
)

// KeyValueField is a single value of a [KeyValues]. Value is canonical text: int is base 10 and bool is true or
// false. ID is stable within the secret, so that renaming the key keeps the stored value.
type KeyValueField struct {
	ID             string          `json:"id"`
	Key            string          `json:"key"`
	Label          string          `json:"label,omitempty"`
	SupportingText string          `json:"supportingText,omitempty"`
	Kind           FieldKind       `json:"kind"`
	Protection     FieldProtection `json:"protection,omitempty"`
	Value          string          `json:"value,omitempty"`
}

// KeyValueSection groups the fields of a [KeyValues] for the user interface.
type KeyValueSection struct {
	ID             string          `json:"id"`
	Label          string          `json:"label,omitempty"`
	SupportingText string          `json:"supportingText,omitempty"`
	Fields         []KeyValueField `json:"fields,omitempty"`
}

// KeyValues is a secret, whose fields are defined by its owner at any time. Some keys have a fixed meaning for a
// connection, see [KeyBaseURL].
type KeyValues struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	LogoSVG     string            `json:"logoSvg,omitempty"`
	Sections    []KeyValueSection `json:"sections,omitempty"`
	_           struct{}          `credentialName:"Schlüssel und Werte" credentialDescription:"Frei definierbare Felder, z. B. die Adresse und die Zugangsdaten einer API, die Programme als Verbindung verwenden oder lesen können."`
}

func (KeyValues) Credentials() bool {
	return true
}

func (kv KeyValues) GetName() string {
	return kv.Name
}

func (kv KeyValues) IsZero() bool {
	return kv.Name == "" && kv.Description == "" && kv.LogoSVG == "" && len(kv.Sections) == 0
}

// Fields iterates over the fields of all sections.
func (kv KeyValues) Fields() iter.Seq[KeyValueField] {
	return func(yield func(KeyValueField) bool) {
		for _, section := range kv.Sections {
			for _, field := range section.Fields {
				if !yield(field) {
					return
				}
			}
		}
	}
}

// Clone returns a deep copy. A nil slice stays nil, so that the copy is deeply equal.
func (kv KeyValues) Clone() KeyValues {
	kv.Sections = slices.Clone(kv.Sections)
	for i := range kv.Sections {
		kv.Sections[i].Fields = slices.Clone(kv.Sections[i].Fields)
	}

	return kv
}

// WithIDs returns a copy, in which every section and field has an identifier.
func (kv KeyValues) WithIDs() KeyValues {
	kv = kv.Clone()
	for i := range kv.Sections {
		section := &kv.Sections[i]
		if section.ID == "" {
			section.ID = data.RandIdent[string]()
		}

		for j := range section.Fields {
			if section.Fields[j].ID == "" {
				section.Fields[j].ID = data.RandIdent[string]()
			}
		}
	}

	return kv
}

// AddField returns a copy with the field at the end of the first section with the given label, which is created if
// it does not exist. Leave the ID of a new field empty. Together with [WithoutWriteOnly] and [UpdateMyCredentials]
// everything else is kept:
//
//	current := secret.WithoutWriteOnly(stored).(secret.KeyValues)
//	err := updateMyCredentials(subject, id, current.AddField("API", field))
func (kv KeyValues) AddField(sectionLabel string, f KeyValueField) KeyValues {
	kv = kv.Clone()
	for i := range kv.Sections {
		if kv.Sections[i].Label == sectionLabel {
			kv.Sections[i].Fields = append(kv.Sections[i].Fields, f)
			return kv
		}
	}

	kv.Sections = append(kv.Sections, KeyValueSection{Label: sectionLabel, Fields: []KeyValueField{f}})
	return kv
}

// WriteOnlyFields implements [DynamicWriteOnly]. The result is keyed by [KeyValueField.ID].
func (kv KeyValues) WriteOnlyFields() map[string]bool {
	res := map[string]bool{}
	for field := range kv.Fields() {
		if field.Protection == WriteOnly {
			res[field.ID] = field.Value != ""
		}
	}

	return res
}

// RedactWriteOnly implements [DynamicWriteOnly].
func (kv KeyValues) RedactWriteOnly() Credentials {
	kv = kv.Clone()
	for i := range kv.Sections {
		for j := range kv.Sections[i].Fields {
			if kv.Sections[i].Fields[j].Protection == WriteOnly {
				kv.Sections[i].Fields[j].Value = ""
			}
		}
	}

	return kv
}

// KeepWriteOnly implements [DynamicWriteOnly]. An empty write-only field gets the stored value of the write-only
// field with the same ID. Thus, a field which has been write-only and is not anymore has lost its stored value. If
// the value of [KeyBaseURL] changes, nothing is carried over. It is compared by the key and not by the ID of its
// field, so that renaming, deleting and adding a field cannot redirect a stored token to another server.
func (kv KeyValues) KeepWriteOnly(stored Credentials) Credentials {
	old, ok := stored.(KeyValues)
	if !ok || old.BaseURL() != kv.BaseURL() {
		return kv
	}

	values := map[string]string{}
	for field := range old.Fields() {
		if field.Protection == WriteOnly && field.ID != "" {
			values[field.ID] = field.Value
		}
	}

	kv = kv.Clone()
	for i := range kv.Sections {
		for j := range kv.Sections[i].Fields {
			field := &kv.Sections[i].Fields[j]
			if field.Protection == WriteOnly && field.Value == "" {
				field.Value = values[field.ID]
			}
		}
	}

	return kv
}

// BaseURL returns the value of [KeyBaseURL], which is empty without such a field.
func (kv KeyValues) BaseURL() string {
	for field := range kv.Fields() {
		if field.Key == KeyBaseURL {
			return field.Value
		}
	}

	return ""
}

// KeyValueInfo describes a field without exposing what is protected, e.g. for a language model.
type KeyValueInfo struct {
	Section        string          `json:"section,omitempty"`
	Key            string          `json:"key"`
	Label          string          `json:"label,omitempty"`
	SupportingText string          `json:"supportingText,omitempty"`
	Kind           FieldKind       `json:"kind"`
	Protection     FieldProtection `json:"protection,omitempty"`
	HasValue       bool            `json:"hasValue"`
	// Value is only set for [Visible] fields.
	Value string `json:"value,omitempty"`
}

// DescribeKeyValues describes all fields and contains only the values of the [Visible] ones.
func DescribeKeyValues(kv KeyValues) []KeyValueInfo {
	var res []KeyValueInfo
	for _, section := range kv.Sections {
		for _, field := range section.Fields {
			info := KeyValueInfo{
				Section:        section.Label,
				Key:            field.Key,
				Label:          field.Label,
				SupportingText: field.SupportingText,
				Kind:           field.Kind,
				Protection:     field.Protection,
				HasValue:       field.Value != "",
			}

			if field.Protection == Visible {
				info.Value = field.Value
			}

			res = append(res, info)
		}
	}

	return res
}

// IsFixedKey returns true, if a connection consumes the key, see [KeyBaseURL]. A program never sees such a value.
func IsFixedKey(key string) bool {
	return key == KeyBaseURL || strings.HasPrefix(key, KeyPrefixHeader) || strings.HasPrefix(key, KeyPrefixQuery)
}

// HeaderForKey returns the name of the header of a key like HEADER_X_API_KEY. It returns false for any other key,
// for the keys of the Authorization header and if a connection must not set the header.
func HeaderForKey(key string) (string, bool) {
	name, ok := strings.CutPrefix(key, KeyPrefixHeader)
	if !ok || name == "" || key == KeyHeaderBearerAuth || key == KeyHeaderBasicAuthUser || key == KeyHeaderBasicAuthPassword {
		return "", false
	}

	header := http.CanonicalHeaderKey(strings.ReplaceAll(name, "_", "-"))
	if !httpguts.ValidHeaderFieldName(header) || forbiddenHeader(header) {
		return "", false
	}

	return header, true
}

// QueryForKey returns the name of the query parameter of a key like QUERY_api_key.
func QueryForKey(key string) (string, bool) {
	name, ok := strings.CutPrefix(key, KeyPrefixQuery)
	return name, ok && name != ""
}

// forbiddenHeaders mirrors the headers, which package nethost forbids a program, because a connection must not
// control the transport either. Only Authorization is allowed here, as it is the very purpose of a connection.
var forbiddenHeaders = map[string]bool{
	"Cookie": true, "Set-Cookie": true, "Host": true, "Connection": true, "Content-Length": true,
	"Transfer-Encoding": true, "Te": true, "Trailer": true, "Upgrade": true,
}

func forbiddenHeader(canonicalName string) bool {
	return forbiddenHeaders[canonicalName] || strings.HasPrefix(canonicalName, "Proxy-") || strings.HasPrefix(canonicalName, "X-Forwarded-")
}
