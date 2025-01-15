package user

import (
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/image"
	"go.wdy.de/nago/pkg/data"
	"golang.org/x/text/language"
	"regexp"
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
	Firstname         string
	Lastname          string
	Email             Email
	Password          Password
	PasswordRepeated  Password
	PreferredLanguage language.Tag
	NotifyUser        bool
	Verified          bool
}

// ID of a user entity in the Nago IAM.
type ID string

type Contact struct {
	Avatar image.ID `json:"avatar,omitempty"`
	// AcademicDegree is e.g. Diploma, Bachelor, Master or Doctor
	AcademicDegree string `json:"academicDegree,omitempty"`
	// OfficialTitle is like Professor, Oberbürgermeister etc.
	OfficialTitle string `json:"officialTitle,omitempty"`
	// Saluation is like Mr, Mrs or divers
	Salutation  string `json:"salutation,omitempty"`
	Firstname   string `json:"firstname,omitempty"`
	Lastname    string `json:"lastname,omitempty"`
	Phone       string `json:"phone,omitempty"`
	MobilePhone string `json:"mobilePhone,omitempty"`
	// Country is like Deutschland, not the BCP47 code
	Country    string `json:"country,omitempty"`
	City       string `json:"city,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	LinkedIn   string `json:"linkedIn,omitempty"`
	Website    string `json:"website,omitempty"`
	// Position is like CEO
	Position    string `json:"position,omitempty"`
	CompanyName string `json:"company,omitempty"`
	// PreferredLanguage is a BCP47 string like de or en_US
	PreferredLanguage string `json:"preferredLanguage,omitempty"`
}

func (d Contact) IsZero() bool {
	return d == Contact{}
}

type Code struct {
	Value      string    `json:"value,omitempty"`
	ValidUntil time.Time `json:"validUntil,omitempty"`
}

func NewCode(lifetime time.Duration) Code {
	return Code{
		Value:      data.RandIdent[string]()[:8],
		ValidUntil: time.Now().Add(lifetime),
	}
}

func (c Code) IsZero() bool {
	return c.Value == "" && c.ValidUntil.IsZero()
}

type User struct {
	ID                    ID              `json:"id"`
	Email                 Email           `json:"email"`
	Contact               Contact         `json:"contact,omitzero"`
	Salt                  []byte          `json:"salt,omitempty"`
	Algorithm             HashAlgorithm   `json:"algorithm,omitempty"`
	PasswordHash          []byte          `json:"passwordHash,omitempty"`
	LastPasswordChangedAt time.Time       `json:"lastPasswordChangedAt"`
	CreatedAt             time.Time       `json:"createdAt"`
	EMailVerified         bool            `json:"emailVerified,omitempty"`
	Status                AccountStatus   `json:"status,omitempty"`
	Roles                 []role.ID       `json:"roles,omitempty"`       // roles may also contain inherited permissions
	Groups                []group.ID      `json:"groups,omitempty"`      // groups may also contain inherited permissions
	Permissions           []permission.ID `json:"permissions,omitempty"` // individual custom permissions
	Licenses              []license.ID    `json:"licenses,omitempty"`
	RequirePasswordChange bool            `json:"requirePasswordChange,omitempty"`
	VerificationCode      Code            `json:"verificationCode,omitzero"`
	PasswordRequestCode   Code            `json:"passwordRequestCode,omitzero"`
}

func (u User) String() string {
	return string(u.ID)
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
