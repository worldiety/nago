package iam

import (
	"bytes"
	"crypto/rand"
	"fmt"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"golang.org/x/crypto/argon2"
	"log/slog"
	rand2 "math/rand"
	"regexp"
	"strings"
	"time"
)

type Email string

var regexMail = regexp.MustCompile(`^[\w-.]+@([\w-]+\.)+[\w-]{2,4}$`)

const maxPasswordLength = 1000

// Valid checks if the Mail looks like structural valid mail. It does not mean that the address actually exists
// or has been verified.
func (e Email) Valid() bool {
	// see https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html#email-address-validation
	if len(e) > 254 {
		return false
	}

	if e == "admin@localhost" {
		return true
	}

	return regexMail.FindString(string(e)) == string(e)
}

func (e Email) Equals(other Email) bool {
	return strings.EqualFold(string(other), string(e))
}

type UserRepository = data.Repository[User, auth.UID]

// #[go.TaggedUnion]
type _AccountStatus interface {
	Enabled | Disabled | EnabledUntil
}

type Enabled struct{}

type Disabled struct{}
type EnabledUntil struct {
	ValidUntil time.Time
}

type PID = string

type HashAlgorithm string

const (
	Argon2IdMin HashAlgorithm = "argon2-id-min"
)

type User struct {
	ID            auth.UID      `json:"ID,omitempty"`
	Email         Email         `json:"email,omitempty"`
	Firstname     string        `json:"firstname,omitempty"`
	Lastname      string        `json:"lastname,omitempty"`
	Salt          []byte        `json:"salt,omitempty"`
	Algorithm     HashAlgorithm `json:"algorithm,omitempty"`
	PasswordHash  []byte        `json:"passwordHash,omitempty"`
	EMailVerified bool          `json:"emailVerified,omitempty"`
	Status        AccountStatus `json:"status,omitempty"`
	Roles         []auth.RID    `json:"roles,omitempty"`       // roles may also contain inherited permissions
	Groups        []auth.GID    `json:"groups,omitempty"`      // groups may also contain inherited permissions
	Permissions   []PID         `json:"permissions,omitempty"` // individual custom permissions
}

func (u User) Identity() auth.UID {
	return u.ID
}

// not safe to export
func (s *Service) userByMail(email Email) (std.Option[User], error) {
	lcMail := Email(strings.ToLower(string(email)))
	if !lcMail.Valid() {
		return std.None[User](), fmt.Errorf("invalid email: %v", email)
	}

	found := false
	var usr User
	var err error
	s.users.Each(func(user User, e error) bool {
		if e != nil {
			err = e
			return false
		}

		if user.Email == lcMail {
			found = true
			usr = user
			return false
		}

		return true
	})

	if err != nil {
		return std.None[User](), err
	}

	if !found {
		return std.None[User](), nil
	}

	return std.Some(usr), nil
}

// authenticates checks if the given mail and password can be authenticated against the email login identifier.
func (s *Service) authenticates(email, password string) (User, error) {
	// see https://owasp.org/www-community/controls/Blocking_Brute_Force_Attacks
	// and mitigate time based attacks for non-constant operations.
	time.Sleep(min(200, time.Duration(rand2.Intn(1000))))

	// see https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#password-storage-cheat-sheet
	// and https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#compare-password-hashes-using-safe-functions
	if len(password) > maxPasswordLength {
		return User{}, fmt.Errorf("password must be less than 1000 characters") // probably a DOS attack
	}

	if !Email(email).Valid() {
		return User{}, fmt.Errorf("invalid email")
	}

	optUsr, err := s.userByMail(Email(email))
	if err != nil {
		return User{}, err
	}

	// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#authentication-responses
	const msg = "user does not exist, or account is disabled, or password is wrong"
	if !optUsr.Valid {
		return User{}, fmt.Errorf(msg)
	}

	usr := optUsr.Unwrap()
	var hash []byte
	switch usr.Algorithm {
	case Argon2IdMin:
		hash = argon2idMin(password, optUsr.Unwrap().Salt)
	default:
		slog.Error("unsupported hash algorithm", "algorithm", usr.Algorithm, "uid", usr.ID)
		return User{}, fmt.Errorf(msg)
	}

	// again, keep code order to protect against time-based early exit attacks
	enabled := MatchAccountStatus(usr.Status,
		func(enabled Enabled) bool {
			return true
		}, func(disabled Disabled) bool {
			return false
		}, func(until EnabledUntil) bool {
			return time.Now().Before(until.ValidUntil)
		}, func(a any) bool {
			return false
		})

	if !enabled {
		return User{}, fmt.Errorf(msg)
	}

	// finally perform the comparison, which is not time-constant, but we have randomized the entire processing
	// time, so this should not be a problem anymore.
	if bytes.Equal(hash, usr.PasswordHash) {
		return usr, nil
	}

	return User{}, fmt.Errorf(msg)
}

func (s *Service) DeleteUser(subject auth.Subject, id auth.UID) error {
	if err := subject.Audit(DeleteUser); err != nil {
		return err
	}

	if err := s.users.DeleteByID(id); err != nil {
		return fmt.Errorf("cannot delete user: %w", err)
	}

	// we do not delete stale foreign keys, which is checked whenever a subject is requested from us, see [Service.Subject]

	// TODO but how to refresh the stored snapshot subject within each window? => create a renew cycle within the event loop

	return nil
}

// this is used in a massive hosting environment, we cannot afford the RFC settings.
// Therefore, we use the following minimal OWASP settings, see https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html:
//
//	Use Argon2id with a minimum configuration of 19 MiB of memory, an iteration count of 2, and 1 degree of parallelism.
func argon2idMin(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 2, 19*1024, 1, 32)
}

func setArgon2idMin(usr *User, password string) error {
	// see https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id
	var salt [32]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return fmt.Errorf("no secure random entropy: %w", err)
	}
	hash := argon2idMin(password, salt[:])

	usr.Salt = salt[:]
	usr.PasswordHash = hash
	usr.Algorithm = Argon2IdMin

	return nil
}

func (s *Service) NewUser(subject auth.Subject, email, firstname, lastname, password string) (auth.UID, error) {
	if err := subject.Audit(CreateUser); err != nil {
		return "", err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock() // this is really harsh and allows intentionally only to create one user per second

	// see https://owasp.org/www-community/controls/Blocking_Brute_Force_Attacks
	// and mitigate time based attacks
	time.Sleep(min(200, time.Duration(rand2.Intn(1000))))

	mail := Email(strings.ToLower(email))
	if !mail.Valid() {
		return "", fmt.Errorf("invalid email: %v", email)
	}

	// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#implement-proper-password-strength-controls
	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}
	const minEntropyBits = 60
	if err := passwordvalidator.Validate(password, minEntropyBits); err != nil {
		return "", fmt.Errorf("password has not enough entropy: %v", err)
	}

	// see https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#password-storage-cheat-sheet
	// and https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#compare-password-hashes-using-safe-functions
	if len(password) > maxPasswordLength {
		return "", fmt.Errorf("password must be less than 1000 characters") // probably a DOS attack
	}

	// see https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id
	var salt [32]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return "", fmt.Errorf("no secure random entropy: %v", err)
	}
	hash := argon2idMin(password, salt[:])

	user := User{
		// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#user-ids
		ID:           data.RandIdent[auth.UID](),
		Email:        mail,
		Algorithm:    Argon2IdMin,
		Firstname:    firstname,
		Lastname:     lastname,
		Salt:         salt[:],
		PasswordHash: hash,
		Status:       AccountStatus{}.WithEnabled(Enabled{}),
	}

	// intentionally validate now, so that an attacker cannot use this method to massively
	// find out, which mails exist in the system
	usr, err := s.userByMail(mail)
	if err != nil {
		return "", fmt.Errorf("cannot check for existing user: %w", err)
	}

	if usr.Valid {
		return "", fmt.Errorf("user email already taken")
	}

	// unlikely, but better safe than sorry
	usr, err = s.users.FindByID(user.ID)
	if err != nil {
		return "", fmt.Errorf("cannot find user by id: %w", err)
	}

	if usr.Valid {
		return "", fmt.Errorf("user id already taken")
	}

	// persist
	err = s.users.Save(user)
	if err != nil {
		return "", fmt.Errorf("cannot persist new user: %w", err)
	}

	return user.ID, nil
}

// TODO ChangeMyPassword and Sensitive/ Priviledged property from LastAuthenticated
