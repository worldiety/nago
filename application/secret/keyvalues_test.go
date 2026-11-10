// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/worldiety/enum"
	"go.wdy.de/nago/pkg/std"
)

var (
	_ = enum.Variant[Credentials, KeyValues](enum.Rename[KeyValues]("nago.secret.keyvalues"))
	_ = enum.Variant[Credentials, addressCredentials](enum.Rename[addressCredentials]("nago.secret.test.address"))
)

// addressCredentials has a write-only value and declares its address by its tags, like [SMTP].
type addressCredentials struct {
	Name    string `json:"name"`
	BaseURL string `json:"baseUrl" credentialAddress:"true"`
	Auth    string `json:"auth"`
	Token   string `json:"token" credentialWriteOnly:"true"`
}

func (c addressCredentials) GetName() string   { return c.Name }
func (c addressCredentials) Credentials() bool { return true }
func (c addressCredentials) IsZero() bool      { return c == addressCredentials{} }

const (
	testURL   = "https://api.example.com/v1"
	otherURL  = "https://evil.example.com"
	testToken = "stored-token"
)

func kvOf(fields ...KeyValueField) KeyValues {
	return KeyValues{Name: "api", Sections: []KeyValueSection{{ID: "s1", Fields: fields}}}
}

func baseURLField(id, value string) KeyValueField {
	return KeyValueField{ID: id, Key: KeyBaseURL, Kind: FieldURL, Value: value}
}

func tokenField(value string) KeyValueField {
	return KeyValueField{ID: "t", Key: KeyHeaderBearerAuth, Kind: FieldString, Protection: WriteOnly, Value: value}
}

func storedToken(c Credentials) string {
	for field := range c.(KeyValues).Fields() {
		if field.ID == "t" {
			return field.Value
		}
	}

	return "<no field>"
}

// A stored token must never follow a changed address, otherwise anyone with edit rights can read it on an own server.
func TestAddressChangeDropsWriteOnlyOfReflectedTypes(t *testing.T) {
	tests := []struct {
		name            string
		stored, updated Credentials
		want            Credentials
	}{
		{
			"http unchanged keeps",
			addressCredentials{Name: "a", BaseURL: testURL, Token: testToken},
			addressCredentials{Name: "b", BaseURL: testURL},
			addressCredentials{Name: "b", BaseURL: testURL, Token: testToken},
		},
		{
			"http other address drops",
			addressCredentials{Name: "a", BaseURL: testURL, Token: testToken},
			addressCredentials{Name: "a", BaseURL: otherURL},
			addressCredentials{Name: "a", BaseURL: otherURL},
		},
		{
			"http other address with new token",
			addressCredentials{Name: "a", BaseURL: testURL, Token: testToken},
			addressCredentials{Name: "a", BaseURL: otherURL, Token: "new"},
			addressCredentials{Name: "a", BaseURL: otherURL, Token: "new"},
		},
		{
			"smtp unchanged keeps",
			SMTP{Host: "smtp.example.com", Port: 587, Username: "a", Password: testToken},
			SMTP{Host: "smtp.example.com", Port: 587, Username: "b"},
			SMTP{Host: "smtp.example.com", Port: 587, Username: "b", Password: testToken},
		},
		{
			"smtp other host drops",
			SMTP{Host: "smtp.example.com", Port: 587, Password: testToken},
			SMTP{Host: "evil.example.com", Port: 587},
			SMTP{Host: "evil.example.com", Port: 587},
		},
		{
			"smtp other port drops",
			SMTP{Host: "smtp.example.com", Port: 587, Password: testToken},
			SMTP{Host: "smtp.example.com", Port: 2525},
			SMTP{Host: "smtp.example.com", Port: 2525},
		},
		{
			"a type without address keeps",
			writeOnlyCredentials{Name: "a", Password: testToken},
			writeOnlyCredentials{Name: "b"},
			writeOnlyCredentials{Name: "b", Password: testToken},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keepWriteOnly(tt.stored, tt.updated); got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAddressChangeDropsWriteOnlyOfKeyValues(t *testing.T) {
	stored := kvOf(baseURLField("b", testURL), tokenField(testToken))
	tests := []struct {
		name    string
		stored  KeyValues
		updated KeyValues
		want    string
	}{
		{"same address keeps", stored, kvOf(baseURLField("b", testURL), tokenField("")), testToken},
		{"other address drops", stored, kvOf(baseURLField("b", otherURL), tokenField("")), ""},
		{"other address keeps a new value", stored, kvOf(baseURLField("b", otherURL), tokenField("new")), "new"},
		{"same address replaces the value", stored, kvOf(baseURLField("b", testURL), tokenField("new")), "new"},
		{
			"renaming the key of the address drops",
			stored,
			kvOf(KeyValueField{ID: "b", Key: "OLD_URL", Kind: FieldURL, Value: testURL}, tokenField("")),
			"",
		},
		{"deleting the address drops", stored, kvOf(tokenField("")), ""},
		{"deleting and adding another address drops", stored, kvOf(baseURLField("new", otherURL), tokenField("")), ""},
		{"a new field with the same address keeps", stored, kvOf(baseURLField("new", testURL), tokenField("")), testToken},
		{"adding an address drops", kvOf(tokenField(testToken)), kvOf(baseURLField("new", otherURL), tokenField("")), ""},
		{
			"renaming another field to the address drops",
			kvOf(KeyValueField{ID: "h", Key: "HOST", Kind: FieldURL, Value: testURL}, tokenField(testToken)),
			kvOf(baseURLField("h", testURL), tokenField("")),
			"",
		},
		{"without address keeps", kvOf(tokenField(testToken)), kvOf(tokenField("")), testToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keepWriteOnly(tt.stored, tt.updated)
			if token := storedToken(got); token != tt.want {
				t.Fatalf("got %q, want %q", token, tt.want)
			}
		})
	}
}

func TestWriteOnlyIsOneWay(t *testing.T) {
	stored := kvOf(tokenField(testToken))
	field := func(protection FieldProtection, value string) KeyValues {
		f := tokenField(value)
		f.Protection = protection
		return kvOf(f)
	}

	tests := []struct {
		name    string
		updated KeyValues
		want    string
	}{
		{"staying write-only keeps", field(WriteOnly, ""), testToken},
		{"concealed loses", field(Concealed, ""), ""},
		{"visible loses", field(Visible, ""), ""},
		{"concealed with a new value", field(Concealed, "new"), "new"},
		{"visible with a new value", field(Visible, "new"), "new"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if token := storedToken(keepWriteOnly(stored, tt.updated)); token != tt.want {
				t.Fatalf("got %q, want %q", token, tt.want)
			}
		})
	}

	// the other direction: an empty field must not pick up a value, which was readable before
	readable := kvOf(KeyValueField{ID: "t", Key: "token", Kind: FieldString, Value: "visible"})
	got := keepWriteOnly(readable, kvOf(KeyValueField{ID: "t", Key: "token", Kind: FieldString, Protection: WriteOnly}))
	if token := storedToken(got); token != "" {
		t.Fatalf("got %q", token)
	}
}

func TestFieldsAreMatchedByID(t *testing.T) {
	stored := KeyValues{Name: "api", Sections: []KeyValueSection{
		{ID: "s1", Fields: []KeyValueField{tokenField(testToken)}},
		{ID: "s2"},
	}}

	tests := []struct {
		name    string
		updated KeyValues
		want    string
	}{
		{
			"a renamed key keeps",
			kvOf(KeyValueField{ID: "t", Key: "renamed", Kind: FieldMultiLine, Protection: WriteOnly}),
			testToken,
		},
		{
			"another section keeps",
			KeyValues{Name: "api", Sections: []KeyValueSection{{ID: "s2", Fields: []KeyValueField{tokenField("")}}}},
			testToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if token := storedToken(keepWriteOnly(stored, tt.updated)); token != tt.want {
				t.Fatalf("got %q, want %q", token, tt.want)
			}
		})
	}

	otherID := kvOf(KeyValueField{ID: "other", Key: KeyHeaderBearerAuth, Kind: FieldString, Protection: WriteOnly})
	if value := keepWriteOnly(stored, otherID).(KeyValues).Sections[0].Fields[0].Value; value != "" {
		t.Fatalf("the same key with another id must not match, got %q", value)
	}

	// a field without an identifier is new, also if the stored one has none
	noID := kvOf(KeyValueField{Key: "a", Kind: FieldString, Protection: WriteOnly, Value: testToken})
	merged := keepWriteOnly(noID, kvOf(KeyValueField{Key: "a", Kind: FieldString, Protection: WriteOnly})).(KeyValues)
	if merged.Sections[0].Fields[0].Value != "" {
		t.Fatal("a field without an identifier must not match")
	}

	res, err := KeyValues{Name: "api", Sections: []KeyValueSection{{Fields: []KeyValueField{{Key: "a", Kind: FieldString}}}}}.Validated()
	if err != nil {
		t.Fatal(err)
	}

	section := res.(KeyValues).Sections[0]
	if section.ID == "" || section.Fields[0].ID == "" {
		t.Fatalf("expected the missing identifiers to be filled: %+v", section)
	}
}

func TestValidated(t *testing.T) {
	const value = "s3cr3t-value"
	long := strings.Repeat("a", MaxLogoSVGBytes)
	field := func(key string, kind FieldKind, protection FieldProtection, v string) KeyValues {
		return kvOf(KeyValueField{ID: "f", Key: key, Kind: kind, Protection: protection, Value: v})
	}

	tests := []struct {
		name  string
		kv    KeyValues
		valid bool
	}{
		{"empty secret with a name", KeyValues{Name: "a"}, true},
		{"no name", KeyValues{}, false},
		{"blank name", KeyValues{Name: "  "}, false},

		{"plain key", field("tenant", FieldString, Visible, value), true},
		{"key with all allowed characters", field("a0_b.c-d", FieldString, Visible, value), true},
		{"empty key", field("", FieldString, Visible, value), false},
		{"key starts with digit", field("0a", FieldString, Visible, value), false},
		{"key starts with underscore", field("_a", FieldString, Visible, value), false},
		{"key with space", field("a b", FieldString, Visible, value), false},
		{"key with newline", field("a\n", FieldString, Visible, value), false},
		{"key with umlaut", field("ä", FieldString, Visible, value), false},
		{"duplicate key in a section", kvOf(KeyValueField{ID: "a", Key: "x", Kind: FieldString}, KeyValueField{ID: "b", Key: "x", Kind: FieldString}), false},
		{"duplicate key across sections", KeyValues{Name: "a", Sections: []KeyValueSection{
			{ID: "s1", Fields: []KeyValueField{{ID: "a", Key: "x", Kind: FieldString}}},
			{ID: "s2", Fields: []KeyValueField{{ID: "b", Key: "x", Kind: FieldString}}},
		}}, false},
		{"duplicate field id", kvOf(KeyValueField{ID: "a", Key: "x", Kind: FieldString}, KeyValueField{ID: "a", Key: "y", Kind: FieldString}), false},
		{"duplicate section id", KeyValues{Name: "a", Sections: []KeyValueSection{{ID: "s"}, {ID: "s"}}}, false},
		{"section and field may share an id", KeyValues{Name: "a", Sections: []KeyValueSection{{ID: "1", Fields: []KeyValueField{{ID: "1", Key: "x", Kind: FieldString}}}}}, true},

		{"unknown kind", field("a", "float", Visible, value), false},
		{"missing kind", field("a", "", Visible, value), false},
		{"unknown protection", field("a", FieldString, "secret", value), false},

		{"write-only string", field("a", FieldString, WriteOnly, value), true},
		{"write-only multi-line", field("a", FieldMultiLine, WriteOnly, value), true},
		{"write-only int", field("a", FieldInt, WriteOnly, "1"), false},
		{"write-only bool", field("a", FieldBool, WriteOnly, "true"), false},
		{"write-only url", field("a", FieldURL, WriteOnly, "https://a.example"), false},
		{"concealed int", field("a", FieldInt, Concealed, "1"), true},

		{"string with newline", field("a", FieldString, Visible, value+"\n"), false},
		{"multi-line", field("a", FieldMultiLine, Visible, value+"\n"+value), true},
		{"int", field("a", FieldInt, Visible, "-12"), true},
		{"int empty", field("a", FieldInt, Visible, ""), true},
		{"int text", field("a", FieldInt, Visible, value), false},
		{"int not canonical", field("a", FieldInt, Visible, "+1"), false},
		{"int too large", field("a", FieldInt, Visible, "99999999999999999999"), false},
		{"bool true", field("a", FieldBool, Visible, "true"), true},
		{"bool false", field("a", FieldBool, Visible, "false"), true},
		{"bool not canonical", field("a", FieldBool, Visible, "1"), false},
		{"url", field("a", FieldURL, Visible, "ftp://a.example/x"), true},
		{"url opaque", field("a", FieldURL, Visible, "mailto:a@example.com"), true},
		{"url without scheme", field("a", FieldURL, Visible, value), false},

		{"base url", field(KeyBaseURL, FieldURL, Visible, testURL), true},
		{"base url concealed", field(KeyBaseURL, FieldURL, Concealed, testURL), true},
		{"base url empty", field(KeyBaseURL, FieldURL, Visible, ""), true},
		{"base url kind", field(KeyBaseURL, FieldString, Visible, testURL), false},
		{"base url write-only", field(KeyBaseURL, FieldURL, WriteOnly, testURL), false},
		{"base url scheme", field(KeyBaseURL, FieldURL, Visible, "ftp://a.example"), false},
		{"base url no host", field(KeyBaseURL, FieldURL, Visible, "https://"), false},
		{"base url only port", field(KeyBaseURL, FieldURL, Visible, "https://:8080"), false},
		{"base url userinfo", field(KeyBaseURL, FieldURL, Visible, "https://"+value+":x@a.example"), false},

		{"bearer", field(KeyHeaderBearerAuth, FieldString, WriteOnly, value), true},
		{"bearer with newline", field(KeyHeaderBearerAuth, FieldMultiLine, WriteOnly, value+"\nX: y"), false},
		{"basic password with newline", field(KeyHeaderBasicAuthPassword, FieldMultiLine, WriteOnly, value+"\n"), false},
		{"header", field("HEADER_X_API_KEY", FieldString, Concealed, value), true},
		{"header authorization", field("HEADER_AUTHORIZATION", FieldString, WriteOnly, value), true},
		{"header without name", field("HEADER_", FieldString, Visible, value), false},
		{"header with newline", field("HEADER_X_API_KEY", FieldMultiLine, Visible, value+"\r\nX: y"), false},
		{"header with control character", field("HEADER_X_API_KEY", FieldString, Visible, value+"\x00"), false},
		{"header multi-line kind with one line", field("HEADER_X_API_KEY", FieldMultiLine, Visible, value), true},
		{"query", field("QUERY_api_key", FieldString, Concealed, value), true},
		{"query without name", field("QUERY_", FieldString, Visible, value), false},
		{"query with newline", field("QUERY_a", FieldMultiLine, Visible, value+"\n"), false},

		{"bearer and basic", kvOf(KeyValueField{ID: "a", Key: KeyHeaderBearerAuth, Kind: FieldString}, KeyValueField{ID: "b", Key: KeyHeaderBasicAuthUser, Kind: FieldString}), false},
		{"bearer and authorization", kvOf(KeyValueField{ID: "a", Key: KeyHeaderBearerAuth, Kind: FieldString}, KeyValueField{ID: "b", Key: "HEADER_AUTHORIZATION", Kind: FieldString}), false},
		{"basic user and password", kvOf(KeyValueField{ID: "a", Key: KeyHeaderBasicAuthUser, Kind: FieldString}, KeyValueField{ID: "b", Key: KeyHeaderBasicAuthPassword, Kind: FieldString}), true},
		{"the same header twice", kvOf(KeyValueField{ID: "a", Key: "HEADER_X_KEY", Kind: FieldString}, KeyValueField{ID: "b", Key: "HEADER_X-KEY", Kind: FieldString}), false},

		{"logo", KeyValues{Name: "a", LogoSVG: `<svg xmlns="http://www.w3.org/2000/svg"/>`}, true},
		{"logo with declaration", KeyValues{Name: "a", LogoSVG: `<?xml version="1.0"?><!-- c --><svg></svg>`}, true},
		{"logo with another root", KeyValues{Name: "a", LogoSVG: `<html><svg/></html>`}, false},
		{"logo with upper case root", KeyValues{Name: "a", LogoSVG: `<SVG/>`}, false},
		{"logo with text in front", KeyValues{Name: "a", LogoSVG: `hello<svg/>`}, false},
		{"logo is no xml", KeyValues{Name: "a", LogoSVG: value}, false},
		{"logo at its limit", KeyValues{Name: "a", LogoSVG: "<svg>" + long[:MaxLogoSVGBytes-11] + "</svg>"}, true},
		{"logo too large", KeyValues{Name: "a", LogoSVG: "<svg>" + long + "</svg>"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.kv.Validated()
			if tt.valid && err != nil {
				t.Fatalf("expected to be valid, got %v", err)
			}

			if !tt.valid && err == nil {
				t.Fatal("expected an error")
			}

			if err != nil && strings.Contains(err.Error(), value) {
				t.Fatalf("the error must not contain the value: %v", err)
			}
		})
	}
}

func TestForbiddenHeaderKeys(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"HEADER_X_API_KEY", "X-Api-Key"},
		{"HEADER_ACCEPT", "Accept"},
		{"HEADER_AUTHORIZATION", "Authorization"},
		{"HEADER_x-custom.name", "X-Custom.name"},
		{"HEADER_HOST", ""},
		{"HEADER_COOKIE", ""},
		{"HEADER_SET_COOKIE", ""},
		{"HEADER_CONNECTION", ""},
		{"HEADER_CONTENT_LENGTH", ""},
		{"HEADER_TRANSFER_ENCODING", ""},
		{"HEADER_TE", ""},
		{"HEADER_TRAILER", ""},
		{"HEADER_UPGRADE", ""},
		{"HEADER_PROXY_AUTHORIZATION", ""},
		{"HEADER_PROXY_CONNECTION", ""},
		{"HEADER_X_FORWARDED_FOR", ""},
		{"HEADER_", ""},
		{KeyHeaderBearerAuth, ""},
		{KeyHeaderBasicAuthUser, ""},
		{KeyHeaderBasicAuthPassword, ""},
		{"tenant", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, ok := HeaderForKey(tt.key)
			if got != tt.want || ok != (tt.want != "") {
				t.Fatalf("got %q %v, want %q", got, ok, tt.want)
			}

			if _, err := kvOf(KeyValueField{ID: "a", Key: tt.key, Kind: FieldString}).Validated(); strings.HasPrefix(tt.key, "HEADER_") && tt.want == "" && !isAuthKey(tt.key) && err == nil {
				t.Fatal("expected the key to be refused")
			}
		})
	}
}

func isAuthKey(key string) bool {
	return key == KeyHeaderBearerAuth || isBasicAuthKey(key)
}

func TestFixedKeys(t *testing.T) {
	for key, want := range map[string]bool{
		KeyBaseURL: true, KeyHeaderBearerAuth: true, "HEADER_ACCEPT": true, "QUERY_a": true,
		"tenant": false, "base_url": false, "header_a": false, "BASE_URLS": false,
	} {
		if IsFixedKey(key) != want {
			t.Errorf("%s: expected %v", key, want)
		}
	}

	if name, ok := QueryForKey("QUERY_Api.key_1"); !ok || name != "Api.key_1" {
		t.Fatalf("expected the name verbatim, got %q", name)
	}
}

func TestParseLogoSVGDropsThePrologue(t *testing.T) {
	const svg = `<svg viewBox="0 0 1 1"></svg>`
	comment := "<!-- " + strings.Repeat("x", 1024) + " -->"
	got, err := ParseLogoSVG([]byte("\xef\xbb\xbf<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" + comment + "\n" + svg))
	if err != nil {
		t.Fatal(err)
	}

	if got != svg {
		t.Fatalf("got %q", got)
	}

	if _, err := ParseLogoSVG([]byte(`<?xml version="1.0" encoding="ISO-8859-1"?><svg/>`)); err == nil {
		t.Fatal("expected a foreign encoding to be refused")
	}
}

func TestDescribeKeyValuesHidesValues(t *testing.T) {
	kv := KeyValues{Name: "api", Sections: []KeyValueSection{{ID: "s", Label: "Zugang", Fields: []KeyValueField{
		{ID: "a", Key: "tenant", Label: "Mandant", SupportingText: "Hilfe", Kind: FieldString, Value: "visible"},
		{ID: "b", Key: "pin", Kind: FieldString, Protection: Concealed, Value: "concealed-value"},
		{ID: "c", Key: "token", Kind: FieldMultiLine, Protection: WriteOnly, Value: "writeonly-value"},
		{ID: "d", Key: "empty", Kind: FieldString, Protection: WriteOnly},
	}}}}

	want := []KeyValueInfo{
		{Section: "Zugang", Key: "tenant", Label: "Mandant", SupportingText: "Hilfe", Kind: FieldString, HasValue: true, Value: "visible"},
		{Section: "Zugang", Key: "pin", Kind: FieldString, Protection: Concealed, HasValue: true},
		{Section: "Zugang", Key: "token", Kind: FieldMultiLine, Protection: WriteOnly, HasValue: true},
		{Section: "Zugang", Key: "empty", Kind: FieldString, Protection: WriteOnly},
	}

	got := DescribeKeyValues(kv)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}

	buf := std.Must(json.Marshal(got))
	if strings.Contains(string(buf), "concealed-value") || strings.Contains(string(buf), "writeonly-value") {
		t.Fatal("protected values must not be described")
	}
}

func TestAddFieldIsPure(t *testing.T) {
	fields := make([]KeyValueField, 1, 4) // room to append, which would alias the slice of a shallow copy
	fields[0] = KeyValueField{ID: "a", Key: "a", Kind: FieldString}
	kv := KeyValues{Name: "api", Sections: []KeyValueSection{{ID: "s", Label: "Zugang", Fields: fields}}}
	before := kv.Clone()

	added := kv.AddField("Zugang", KeyValueField{Key: "b", Kind: FieldString})
	if len(added.Sections) != 1 || len(added.Sections[0].Fields) != 2 || added.Sections[0].Fields[1].Key != "b" {
		t.Fatalf("expected the field in the existing section, got %+v", added)
	}

	created := kv.AddField("Neu", KeyValueField{Key: "c", Kind: FieldString})
	if len(created.Sections) != 2 || created.Sections[1].Label != "Neu" || created.Sections[1].Fields[0].Key != "c" {
		t.Fatalf("expected a new section, got %+v", created)
	}

	added.Sections[0].Fields[0].Value = "changed"
	if !reflect.DeepEqual(kv, before) || fields[:2][1] != (KeyValueField{}) {
		t.Fatalf("the original must not be changed, got %+v", kv)
	}
}

func TestWriteOnlyOfKeyValues(t *testing.T) {
	kv := kvOf(
		KeyValueField{ID: "a", Key: "a", Kind: FieldString, Protection: WriteOnly, Value: testToken},
		KeyValueField{ID: "b", Key: "b", Kind: FieldString, Protection: WriteOnly},
		KeyValueField{ID: "c", Key: "c", Kind: FieldString, Protection: Concealed, Value: "c"},
		KeyValueField{ID: "d", Key: "d", Kind: FieldString, Value: "d"},
	)

	if got := WriteOnlyFields(kv); !reflect.DeepEqual(got, map[string]bool{"a": true, "b": false}) {
		t.Fatalf("got %v", got)
	}

	redacted := WithoutWriteOnly(kv).(KeyValues)
	values := []string{}
	for field := range redacted.Fields() {
		values = append(values, field.Value)
	}

	if !reflect.DeepEqual(values, []string{"", "", "c", "d"}) {
		t.Fatalf("got %v", values)
	}

	if kv.Sections[0].Fields[0].Value != testToken {
		t.Fatal("the original must not be changed")
	}
}

func TestKeyValuesIsZero(t *testing.T) {
	if !(KeyValues{}).IsZero() {
		t.Fatal("expected zero")
	}

	for _, kv := range []KeyValues{{Name: "a"}, {Description: "a"}, {LogoSVG: "a"}, {Sections: []KeyValueSection{{}}}} {
		if kv.IsZero() {
			t.Fatalf("%+v must not be zero", kv)
		}
	}
}

// The way of a tool of a language model: everything is kept without ever knowing a protected value.
func TestUpdateWithAddedFieldKeepsEverything(t *testing.T) {
	uc, repo, _ := newTestUseCases(t)
	id, err := uc.CreateSecret(sub(owner), KeyValues{Name: "api", Sections: []KeyValueSection{{Label: "Zugang", Fields: []KeyValueField{
		{Key: KeyBaseURL, Kind: FieldURL, Value: testURL},
		{Key: KeyHeaderBearerAuth, Kind: FieldString, Protection: WriteOnly, Value: testToken},
		{Key: "pin", Kind: FieldString, Protection: Concealed, Value: "1234"},
	}}}})
	if err != nil {
		t.Fatal(err)
	}

	first := mustLoad(t, repo, id).Credentials.(KeyValues)
	for field := range first.Fields() {
		if field.ID == "" {
			t.Fatalf("expected an identifier: %+v", field)
		}
	}

	update := WithoutWriteOnly(first).(KeyValues).AddField("Zugang", KeyValueField{Key: "tenant", Kind: FieldString, Value: "acme"})
	if err := uc.UpdateMyCredentials(sub(owner), id, update); err != nil {
		t.Fatal(err)
	}

	got := mustLoad(t, repo, id).Credentials.(KeyValues)
	if len(got.Sections) != 1 || len(got.Sections[0].Fields) != 4 {
		t.Fatalf("unexpected fields: %+v", got)
	}

	if got.Sections[0].Fields[1].ID != first.Sections[0].Fields[1].ID || got.Sections[0].Fields[1].Value != testToken {
		t.Fatalf("expected the token to be kept: %+v", got.Sections[0].Fields[1])
	}

	if got.Sections[0].Fields[3].ID == "" {
		t.Fatal("expected the new field to get an identifier")
	}
}

func TestUseCasesValidate(t *testing.T) {
	uc, repo, _ := newTestUseCases(t)
	before := std.Must(repo.Count())
	if _, err := uc.CreateSecret(sub(owner), KeyValues{}); err == nil {
		t.Fatal("expected a secret without a name to be refused")
	}

	if std.Must(repo.Count()) != before {
		t.Fatal("expected no secret to be stored")
	}

	id, err := uc.CreateSecret(sub(owner), kvOf(tokenField(testToken)))
	if err != nil {
		t.Fatal(err)
	}

	// validation happens after merging, so a stored value, which does not fit the new key anymore, is refused
	multiLine := kvOf(KeyValueField{ID: "t", Key: "note", Kind: FieldMultiLine, Protection: WriteOnly, Value: "a\nb"})
	if err := uc.UpdateMyCredentials(sub(owner), id, multiLine); err != nil {
		t.Fatal(err)
	}

	renamed := kvOf(KeyValueField{ID: "t", Key: "HEADER_X_KEY", Kind: FieldMultiLine, Protection: WriteOnly})
	err = uc.UpdateMyCredentials(sub(owner), id, renamed)
	if err == nil {
		t.Fatal("expected the merged value to be refused")
	}

	if strings.Contains(err.Error(), "a\nb") {
		t.Fatalf("the error must not contain the value: %v", err)
	}

	if key := mustLoad(t, repo, id).Credentials.(KeyValues).Sections[0].Fields[0].Key; key != "note" {
		t.Fatalf("expected nothing to be stored, got %q", key)
	}
}

// The hole in the use case itself: redirecting the address must not take the stored token along.
func TestUpdateDropsTokenWhenAddressChanges(t *testing.T) {
	uc, repo, _ := newTestUseCases(t)
	id, err := uc.CreateSecret(sub(owner), addressCredentials{Name: "api", BaseURL: testURL, Auth: "bearer", Token: testToken})
	if err != nil {
		t.Fatal(err)
	}

	if err := uc.UpdateMyCredentials(sub(owner), id, addressCredentials{Name: "api", BaseURL: otherURL, Auth: "bearer"}); err != nil {
		t.Fatal(err)
	}

	if got := mustLoad(t, repo, id).Credentials.(addressCredentials); got.BaseURL != otherURL || got.Token != "" {
		t.Fatalf("got %+v", got)
	}
}

func TestHasAddressFields(t *testing.T) {
	if !HasAddressFields(addressCredentials{}) || !HasAddressFields(SMTP{}) || HasAddressFields(writeOnlyCredentials{}) || HasAddressFields(KeyValues{}) {
		t.Fatal("unexpected address fields")
	}
}
