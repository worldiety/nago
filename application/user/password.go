package user

import (
	"bytes"
	"crypto/rand"
	"fmt"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"go.wdy.de/nago/pkg/std"
	"golang.org/x/crypto/argon2"
	rand2 "math/rand"
	"time"
	"unicode/utf8"
)

const (
	minPasswordLength = 8
	minEntropyBits    = 60
	maxPasswordLength = 1000
)

var noLoginErr = std.NewLocalizedError("Login nicht möglich", "Der Nutzer existiert nicht, das Konto ist deaktiviert oder das Passwort ist falsch.")

type Password string

func (p Password) CompareHashAndPassword(algo HashAlgorithm, salt []byte, hash []byte) error {
	// see https://owasp.org/www-community/controls/Blocking_Brute_Force_Attacks
	// and mitigate time based attacks for non-constant operations.
	time.Sleep(min(200, time.Duration(rand2.Intn(1000))))

	// we intentionally validate the password basics, because otherwise a compromised password would open
	// up the attack surface. imagine lifting the password length from 2 chars to 8 chars
	if err := p.Validate(); err != nil {
		return err
	}

	var myHash []byte
	switch algo {
	case Argon2IdMin:
		myHash = argon2idMin(string(p), salt)
	default:
		return fmt.Errorf("unsupported hash algorithm: %s", algo)
	}

	if !bytes.Equal(hash, myHash) {
		// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#authentication-responses
		return noLoginErr // TODO use language from subject
	}

	return nil
}

func (p Password) Validate() error {
	// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#implement-proper-password-strength-controls
	runeCount := utf8.RuneCountInString(string(p))
	if runeCount < minPasswordLength {
		return std.NewLocalizedError("Schwaches Passwort", fmt.Sprintf("Das Kennwort muss mindestens %d Zeichen enthalten.", minPasswordLength)) //TODO use language from subject
	}

	if err := passwordvalidator.Validate(string(p), minEntropyBits); err != nil {
		return std.NewLocalizedError("Schwaches Passwort", "Das Kennwort hat nicht genug Entropie.") // TODO use language from subject
	}

	// see https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#password-storage-cheat-sheet
	// and https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#compare-password-hashes-using-safe-functions
	if runeCount > maxPasswordLength {
		// probably a DOS attack
		return std.NewLocalizedError("Eingabebeschränkung", fmt.Sprintf("Das Kennwort muss kürzer als %d Zeichen sein", maxPasswordLength)) // TODO use language from subject
	}

	return nil
}

func (p Password) Hash(algo HashAlgorithm) (salt []byte, hash []byte, err error) {
	if algo != Argon2IdMin {
		return nil, nil, fmt.Errorf("unsupported hash algorithm")
	}

	// see https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id
	var tmp [32]byte
	if _, err := rand.Read(tmp[:]); err != nil {
		return nil, nil, fmt.Errorf("no secure random entropy: %v", err)
	}

	hash = argon2idMin(string(p), tmp[:])

	return tmp[:], hash, nil
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
