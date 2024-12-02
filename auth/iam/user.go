package iam

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"regexp"
	"slices"
	slices2 "slices"
	"strings"
	"time"
)

type Email string

var regexMail = regexp.MustCompile(`^[\w-.]+@([\w-]+\.)+[\w-]{2,4}$`)

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

	if e == "" {
		return false
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

	for user, err := range s.users.All() {
		if err != nil {
			return std.None[User](), err
		}

		if user.Email == lcMail {
			return std.Some(user), nil
		}
	}

	return std.None[User](), nil
}

// authenticates checks if the given mail and password can be authenticated against the email login identifier.
func (s *Service) authenticates(email string, password Password) (User, error) {

	if !Email(email).Valid() {
		return User{}, std.NewLocalizedError("Login nicht möglich", "Dieses EMail-Adressformat ist nicht erlaubt.")
	}

	optUsr, err := s.userByMail(Email(email))
	if err != nil {
		return User{}, err
	}

	// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#authentication-responses

	if optUsr.IsNone() {
		return User{}, noLoginErr
	}

	usr := optUsr.Unwrap()
	if err := password.CompareHashAndPassword(usr.Algorithm, usr.Salt, usr.PasswordHash); err != nil {
		return User{}, err
	}

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
		return User{}, noLoginErr
	}

	return usr, nil
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

func (s *Service) FindUser(subject auth.Subject, id auth.UID) (std.Option[User], error) {
	if err := subject.Audit(ReadUser); err != nil {
		return std.None[User](), err
	}

	return s.users.FindByID(id)
}

func (s *Service) collectPermissions(usr User) (map[PID]struct{}, error) {
	var err error
	tmp := map[string]struct{}{}
	for _, permission := range usr.Permissions {
		tmp[permission] = struct{}{}
	}

	// collect inherited permissions from roles
	for role, e := range s.roles.FindAllByID(slices2.Values(usr.Roles)) {
		if e != nil {
			err = e
			break
		}

		for _, permission := range role.Permissions {
			tmp[permission] = struct{}{}
		}

	}

	return tmp, err
}

func (s *Service) FindAllUserPermissions(subject auth.Subject, uid auth.UID) ([]PID, error) {
	if err := subject.Audit(ReadUser); err != nil {
		return nil, err
	}

	optUsr, err := s.users.FindByID(uid)
	if err != nil {
		return nil, err
	}

	if !optUsr.Valid {
		return nil, nil
	}

	m, err := s.collectPermissions(optUsr.Unwrap())
	if err != nil {
		return nil, err
	}

	var tmp []PID
	for pid, _ := range m {
		tmp = append(tmp, pid)
	}

	slices.Sort(tmp)
	return tmp, nil
}

func (s *Service) UpdateUser(subject auth.Subject, user auth.UID, email, firstname, lastname string, customPermissions []PID, roles []auth.RID, groups []auth.GID) error {
	if err := subject.Audit(UpdateUser); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock() // this is really harsh and allows intentionally only to create one user per second

	mail := Email(strings.ToLower(email))
	if !mail.Valid() {
		return fmt.Errorf("invalid email: %v", email)
	}

	// intentionally validate now, so that an attacker cannot use this method to massively
	// find out, which mails exist in the system
	optUsr, err := s.userByMail(mail)
	if err != nil {
		return fmt.Errorf("cannot check for existing user: %w", err)
	}

	if optUsr.Valid && optUsr.Unwrap().ID != user {
		return fmt.Errorf("user email already taken")
	}

	optUsr, err = s.users.FindByID(user)
	if err != nil {
		return fmt.Errorf("cannot load existing user: %w", err)
	}

	if !optUsr.Valid {
		return fmt.Errorf("user does not exist")
	}

	usr := optUsr.Unwrap()
	usr.Email = mail
	usr.Firstname = firstname
	usr.Lastname = lastname
	usr.Permissions = customPermissions
	usr.Roles = roles
	usr.Groups = groups

	return s.users.Save(usr)
}

func (s *Service) ChangeMyPassword(subject auth.Subject, oldPassword, newPassword, newRepeated Password) (auth.UID, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock() // this is really harsh and allows intentionally only to change one user per second

	if !subject.Valid() {
		return "", fmt.Errorf("invalid subject")
	}

	if oldPassword == newPassword {
		return "", std.NewLocalizedError("Eingabebeschränkung", "Das alte und das neue Kennwort müssen sich unterscheiden.")
	}

	if newPassword != newRepeated {
		return "", std.NewLocalizedError("Eingabefehler", "Das neue Kennwort unterscheidet sich in der wiederholten Eingabe.")
	}

	if err := newPassword.Validate(); err != nil {
		return "", err
	}

	// check if old password authenticates
	optUsr, err := s.users.FindByID(subject.ID())
	if err != nil {
		return "", fmt.Errorf("cannot find existing user: %w", err)
	}

	if optUsr.IsNone() {
		return "", fmt.Errorf("user has just disappeared")
	}

	usr := optUsr.Unwrap()

	if err := oldPassword.CompareHashAndPassword(Argon2IdMin, usr.Salt, usr.PasswordHash); err != nil {
		return "", err
	}

	// create new credentials
	newSalt, newHash, err := newPassword.Hash(Argon2IdMin)
	if err != nil {
		return "", err
	}

	usr.Salt = newSalt
	usr.PasswordHash = newHash
	usr.Algorithm = Argon2IdMin

	if err := s.users.Save(usr); err != nil {
		return "", fmt.Errorf("cannot update user with new password: %w", err)
	}

	return subject.ID(), nil
}

func (s *Service) NewUser(subject auth.Subject, email, firstname, lastname string, password, passwordRepeat Password) (auth.UID, error) {
	if err := subject.Audit(CreateUser); err != nil {
		return "", err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock() // this is really harsh and allows intentionally only to create one user per second

	if password != passwordRepeat {
		return "", std.NewLocalizedError("Eingabebeschränkung", "Die Kennwörter stimmen nicht überein.")
	}

	mail := Email(strings.ToLower(email))
	if !mail.Valid() {
		return "", fmt.Errorf("invalid email: %v", email)
	}

	if err := password.Validate(); err != nil {
		return "", err
	}

	salt, hash, err := password.Hash(Argon2IdMin)
	if err != nil {
		return "", err
	}

	user := User{
		// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#user-ids
		ID:           data.RandIdent[auth.UID](),
		Email:        mail,
		Algorithm:    Argon2IdMin,
		Firstname:    firstname,
		Lastname:     lastname,
		Salt:         salt,
		PasswordHash: hash,
		Status:       AccountStatus{}.WithEnabled(Enabled{}),
	}

	// intentionally validate now, so that an attacker cannot use this method to massively
	// find out, which mails exist in the system
	usr, err := s.userByMail(mail)
	if err != nil {
		return "", fmt.Errorf("cannot check for existing user: %w", err)
	}

	if usr.IsSome() {
		return "", std.NewLocalizedError("Nutzerregistrierung", "Die EMail-Adresse wird bereits verwendet.")
	}

	// unlikely, but better safe than sorry
	usr, err = s.users.FindByID(user.ID)
	if err != nil {
		return "", fmt.Errorf("cannot find user by id: %w", err)
	}

	if usr.IsSome() {
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
