package usm

import (
	"bytes"
	"crypto/rand"
	"fmt"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"golang.org/x/crypto/argon2"
	rand2 "math/rand"
	"strings"
	"sync"
	"time"
)

const maxPasswordLength = 1000

type UserService struct {
	mutex sync.RWMutex
	repo  UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// not safe to export
func (s *UserService) userByLogin(email auth.EMail) (std.Option[User], error) {
	lcMail := auth.EMail(strings.ToLower(string(email)))
	if !lcMail.Valid() {
		return std.None[User](), fmt.Errorf("invalid email")
	}

	found := false
	var usr User
	var err error
	s.repo.Each(func(user User, e error) bool {
		if e != nil {
			err = e
			return false
		}

		if user.Login == lcMail {
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

func (s *UserService) NewUser(email, firstname, lastname, password string) (auth.UserID, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock() // this is really harsh and allows intentionally only to create one user per second

	// see https://owasp.org/www-community/controls/Blocking_Brute_Force_Attacks
	// and mitigate time based attacks
	time.Sleep(min(200, time.Duration(rand2.Intn(1000))))

	mail := auth.EMail(strings.ToLower(email))
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
	hash := argon2id(password, salt[:])

	user := User{
		// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#user-ids
		ID:           data.RandIdent[auth.UserID](),
		Login:        mail,
		Firstname:    firstname,
		Lastname:     lastname,
		Salt:         salt[:],
		PasswordHash: hash,
	}

	// intentionally validate now, so that an attacker cannot use this method to massively
	// find out, which mails exist in the system
	usr, err := s.userByLogin(mail)
	if err != nil || usr.Valid {
		return "", fmt.Errorf("user email already taken")
	}

	err = s.repo.Save(user)
	if err != nil {
		return "", fmt.Errorf("cannot persist new user: %w", err)
	}

	s.accountCreated(user)

	return user.ID, nil
}

// not safe to export
func (s *UserService) changePassword(id auth.UserID, password string) error {
	optUsr, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if !optUsr.Valid {
		return fmt.Errorf("invalid user")
	}

	usr := optUsr.Unwrap()
	// see https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id
	var salt [32]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return fmt.Errorf("no secure random entropy: %w", err)
	}
	hash := argon2id(password, salt[:])

	usr.Salt = salt[:]
	usr.PasswordHash = hash

	return s.repo.Save(usr)
}

// Authenticates checks if the given mail and password can be authenticated against the email login identifier.
func (s *UserService) Authenticates(email, password string) (bool, error) {
	// see https://owasp.org/www-community/controls/Blocking_Brute_Force_Attacks
	// and mitigate time based attacks for non-constant operations.
	time.Sleep(min(200, time.Duration(rand2.Intn(1000))))

	// see https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#password-storage-cheat-sheet
	// and https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#compare-password-hashes-using-safe-functions
	if len(password) > maxPasswordLength {
		return false, fmt.Errorf("password must be less than 1000 characters") // probably a DOS attack
	}

	if !auth.EMail(email).Valid() {
		return false, fmt.Errorf("invalid email")
	}

	optUsr, err := s.userByLogin(auth.EMail(email))
	if err != nil {
		return false, err
	}

	// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#authentication-responses
	const msg = "user does not exist, or account is disabled, or password is wrong"
	if !optUsr.Valid {
		return false, fmt.Errorf(msg)
	}

	hash := argon2id(password, optUsr.Unwrap().Salt)

	// again, keep code order to protect against time-based early exit attacks
	if optUsr.Unwrap().Disabled {
		return false, fmt.Errorf(msg)
	}

	// finally perform the comparison, which is not time-constant, but we have randomized the entire processing
	// time, so this should not be a problem anymore.
	if bytes.Equal(hash, optUsr.Unwrap().PasswordHash) {
		return true, nil
	}

	return false, fmt.Errorf(msg)
}

// not safe to export
func (s *UserService) listUsers(yield func(User, error) bool) {
	s.repo.Each(yield)
}

// accountCreated may send an email or other domain events
func (s *UserService) accountCreated(user User) {
	//TODO
}

// these parameters are more than the OWASP recommendation and are based on the current RFC recommendations
func argon2id(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
}
