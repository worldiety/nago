// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"regexp"
	"strings"
	"time"

	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/address"
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/pkg/xtime"
	"golang.org/x/text/language"
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

var regexMail = regexp.MustCompile(`^[\w-.]+@([\w-]+\.)+[\w-]{2,20}$`)

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
		c.Addresses.IsZero() &&
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
	NLSManagedUser        bool          `json:"nls,omitempty"`
	VerificationCode      Code          `json:"verificationCode,omitzero"`
	PasswordRequestCode   Code          `json:"passwordRequestCode,omitzero"`

	// some legal/regulatory properties
	Consents []consent.Consent `json:"consents,omitzero"`
}

func (u User) String() string {
	return xstrings.Join2(" ", u.Contact.Firstname, u.Contact.Lastname) + " (" + string(u.Email) + ")"
}

func (u User) SSO() bool {
	return u.NLSManagedUser
}

func (u User) Identity() ID {
	return u.ID
}

func (u User) WithIdentity(id ID) User {
	u.ID = id
	return u
}

func (u User) ConsentStatusByID(id consent.ID) consent.Status {
	for _, c := range u.Consents {
		if c.ID == id {
			return c.Status()
		}
	}

	return consent.Revoked
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
	if u.NLSManagedUser {
		return false
	}

	return u.RequirePasswordChange || !u.PasswordRequestCode.IsZero() || !u.VerificationCode.IsZero() || len(u.PasswordHash) == 0
}
