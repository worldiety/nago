// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/address"
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/pkg/xtime"
	"golang.org/x/text/language"
	"iter"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var AccountStatusEnum = enum.Declare[AccountStatus, func(func(Enabled), func(Disabled), func(EnabledUntil), func(any))]()

type AccountStatus interface {
	accountStatus()
}

type Enabled struct{}

func (Enabled) accountStatus() {}

type Disabled struct{}

func (Disabled) accountStatus() {}

type EnabledUntil struct {
	ValidUntil time.Time
}

func (EnabledUntil) accountStatus() {}

type HashAlgorithm string

const (
	Argon2IdMin HashAlgorithm = "argon2-id-min"
)

var HashAlgorithmValues = []HashAlgorithm{Argon2IdMin}

type Email string

var regexMail = regexp.MustCompile(`^[\w-.]+@([\w-]+\.)+[\w-]{2,4}$`)

// Valid checks if the Mail looks like structural valid mail. It does not mean that the address actually exists
// or has been verified. There are also way more kinds of technical valid addresses, e.g. without a top level
// domain or umlauts, which we may not accept.
func (e Email) Valid() bool {
	// see https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html#email-address-validation
	if len(e) > 254 {
		return false
	}

	if e == "admin@localhost" {
		return true
	}

	if e == "" {
		return false
	}

	return regexMail.FindString(string(e)) == string(e)
}

func (e Email) Equals(other Email) bool {
	return strings.EqualFold(string(other), string(e))
}

type ShortRegistrationUser struct {
	SelfRegistered    bool
	Firstname         string
	Lastname          string
	Email             Email
	Password          Password
	PasswordRepeated  Password
	PreferredLanguage language.Tag
	NotifyUser        bool
	Verified          bool

	Consents []consent.Consent
	// additional optional contact data
	Salutation        string
	Title             string
	Position          string
	CompanyName       string
	City              string
	PostalCode        string
	State             string
	Country           string
	ProfessionalGroup string
	MobilePhone       string
}

// ID of a user entity in the Nago IAM.
type ID string

type Contact struct {
	Avatar image.ID `json:"avatar,omitempty"`
	// Title incorporates the official title like Professor, Oberbürgermeister etc. but also
	// an academic degree like Diploma, Bachelor, Master or Doctor
	Title string `json:"title,omitempty"`
	// Saluation is like Mr, Mrs or divers
	Salutation  string     `json:"salutation,omitempty"`
	Firstname   string     `json:"firstname,omitempty"`
	Lastname    string     `json:"lastname,omitempty"`
	Phone       string     `json:"phone,omitempty"`
	MobilePhone string     `json:"mobilePhone,omitempty"`
	DayOfBirth  xtime.Date `json:"dayOfBirth,omitzero"`

	// Address section
	// Country is like Deutschland, not the BCP47 code
	//deprecated use Addresses
	Country string `json:"country,omitempty"`
	//deprecated use Addresses
	State string `json:"state,omitempty"`
	//deprecated use Addresses
	PostalCode string `json:"postalCode,omitempty"`
	//deprecated use Addresses
	City string `json:"city,omitempty"`

	// These Addresses are owned and part of this address and cannot exist without the enclosing contact.
	Addresses xslices.Slice[address.Address] `json:"addresses,omitempty"`

	// Social Media section
	LinkedIn string `json:"linkedIn,omitempty"`
	Website  string `json:"website,omitempty"`
	// Position is like CEO
	Position          string `json:"position,omitempty"`
	ProfessionalGroup string `json:"professionalGroup,omitempty"`
	CompanyName       string `json:"company,omitempty"`
	// DisplayLanguage is a BCP47 string like de or en_US of what the User wants to see its content.
	DisplayLanguage string `json:"displayLanguage,omitempty"`
	AboutMe         string `json:"aboutMe,omitempty"`
}

func (c Contact) IsZero() bool {
	return c.Avatar == "" &&
		c.Title == "" &&
		c.Salutation == "" &&
		c.Firstname == "" &&
		c.Lastname == "" &&
		c.Phone == "" &&
		c.MobilePhone == "" &&
		c.DayOfBirth.IsZero() &&
		c.Country == "" &&
		c.State == "" &&
		c.PostalCode == "" &&
		c.City == "" &&
		c.IsZero() &&
		c.LinkedIn == "" &&
		c.Website == "" &&
		c.Position == "" &&
		c.ProfessionalGroup == "" &&
		c.CompanyName == "" &&
		c.DisplayLanguage == "" &&
		c.AboutMe == ""
}

type Code struct {
	Value      string    `json:"value,omitempty"`
	ValidUntil time.Time `json:"validUntil,omitempty"`
}

// NewCode returns a code with varying complexity based on the given lifetime.
func NewCode(lifetime time.Duration) Code {
	var code string
	switch {
	case lifetime < time.Minute*5:
		code = data.RandIdent[string]()[:6]
	case lifetime < time.Hour*24:
		code = data.RandIdent[string]()[:8]
	default:
		code = data.RandIdent[string]()
	}

	return Code{
		Value:      code,
		ValidUntil: time.Now().Add(lifetime),
	}
}

func (c Code) IsZero() bool {
	return c.Value == "" && c.ValidUntil.IsZero()
}

type Resource struct {
	// Name of the Store or Repository
	Name string

	// ID is the string version of the root aggregate or entity key used in the named Store or Repository.
	// If ID is empty, all values from the Named Store or Repository are applicable.
	ID string
}

func (r Resource) MarshalText() ([]byte, error) {
	return []byte(strconv.Quote(r.Name + "/" + r.ID)), nil
}

func (r *Resource) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		r.Name = ""
		r.ID = ""
	}

	str, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	tokens := strings.SplitN(str, "/", 2)
	if len(tokens) != 2 {
		return fmt.Errorf("invalid json format for resource: %s", str)
	}

	r.Name = tokens[0]
	r.ID = tokens[1]
	return nil
}

type ResourceWithPermissions struct {
	Permissions iter.Seq[permission.ID]
	Resource
}

type User struct {
	ID                    ID            `json:"id"`
	Email                 Email         `json:"email"`
	Contact               Contact       `json:"contact,omitzero"`
	Salt                  []byte        `json:"salt,omitempty"`
	Algorithm             HashAlgorithm `json:"algorithm,omitempty"`
	PasswordHash          []byte        `json:"passwordHash,omitempty"`
	LastPasswordChangedAt time.Time     `json:"lastPasswordChangedAt"`
	CreatedAt             time.Time     `json:"createdAt"`
	EMailVerified         bool          `json:"emailVerified,omitempty"`
	Status                AccountStatus `json:"status,omitempty"`
	RequirePasswordChange bool          `json:"requirePasswordChange,omitempty"`
	VerificationCode      Code          `json:"verificationCode,omitzero"`
	PasswordRequestCode   Code          `json:"passwordRequestCode,omitzero"`

	// some legal/regulatory properties
	Consents []consent.Consent `json:"consents,omitzero"`

	// global permissions
	Roles       []role.ID       `json:"roles,omitempty"`       // roles may also contain inherited permissions
	Groups      []group.ID      `json:"groups,omitempty"`      // groups may also contain inherited permissions
	Permissions []permission.ID `json:"permissions,omitempty"` // individual custom permissions
	Licenses    []license.ID    `json:"licenses,omitempty"`

	// resource based permissions
	Resources map[Resource][]permission.ID `json:"resources,omitempty"`
}

func (u User) String() string {
	return xstrings.Join2(" ", u.Contact.Firstname, u.Contact.Lastname) + " (" + string(u.Email) + ")"
}

func (u User) Identity() ID {
	return u.ID
}

func (u User) WithIdentity(id ID) User {
	u.ID = id
	return u
}

func (u User) Enabled() bool {
	enabled := false
	AccountStatusEnum.Switch(u.Status)(
		func(Enabled) {
			enabled = true
		},
		func(disabled Disabled) {

		},
		func(until EnabledUntil) {
			enabled = time.Now().Before(until.ValidUntil)
		},
		func(a any) {

		},
	)

	return enabled
}

func (u User) RequiresVerification() bool {
	return u.RequirePasswordChange || !u.PasswordRequestCode.IsZero() || !u.VerificationCode.IsZero() || len(u.PasswordHash) == 0
}
