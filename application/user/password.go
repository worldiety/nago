package user

import (
	"bytes"
	"crypto/rand"
	"fmt"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"go.wdy.de/nago/pkg/std"
	"golang.org/x/crypto/argon2"
	"regexp"
	"unicode/utf8"
)

var noLoginErr = std.NewLocalizedError("Login nicht möglich", "Der Nutzer existiert nicht, das Konto ist deaktiviert oder das Passwort ist falsch.")

type Password string

func (p Password) CompareHashAndPassword(algo HashAlgorithm, salt []byte, hash []byte) error {
	// see https://owasp.org/www-community/controls/Blocking_Brute_Force_Attacks
	// and mitigate time based attacks for non-constant operations.
	// security note: we must not introduce another sleep here, because argon2id should be safe enough
	// and we must not stall any locks here. We need a totally different attack mitigation strategy.

	// security note: we got so many complains, that we don't reject valid but insecure passwords due to raising
	// the passwords requirements. This must be ensured through the user object, by requesting the password
	// change flag.

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
	strength := CalculatePasswordStrength(string(p))
	if !strength.Acceptable {
		return PasswordStrengthError{Strength: strength}
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

type Complexity int

const (
	VeryWeak Complexity = iota
	Weak
	Strong
)

type PasswordStrengthError struct {
	Strength PasswordStrengthIndicator
}

func (e PasswordStrengthError) Error() string {
	return fmt.Sprintf("PasswordStrengthError: %#v", e.Strength)
}

// PasswordStrengthIndicator represents the strength estimation of an analyzed string for the purpose of
// a password. See also BSI recommendations regarding passwords:
// https://www.bsi.bund.de/DE/Themen/Verbraucherinnen-und-Verbraucher/Informationen-und-Empfehlungen/Cyber-Sicherheitsempfehlungen/Accountschutz/Sichere-Passwoerter-erstellen/sichere-passwoerter-erstellen_node.html
// or https://www.bsi.bund.de/SharedDocs/Downloads/DE/BSI/Checklisten/sichere_passwoerter_faktenblatt.pdf?__blob=publicationFile&v=4#download=1
type PasswordStrengthIndicator struct {
	Complexity                Complexity
	ComplexityScale           float64 // 0 = weak, 1 = super strong
	MinLengthRequired         int
	ContainsMinLength         bool
	ContainsSpecial           bool
	ContainsNumber            bool
	ContainsUpperAndLowercase bool
	ContainsBelowMaxLength    bool
	MaxLengthRequired         int
	Acceptable                bool
}

var regexSpecialChar = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]+`)
var regexNumber = regexp.MustCompile(`[0-9]+`)
var regexUppercase = regexp.MustCompile(`[A-ZÄÖÜ]+`)
var regexLowercase = regexp.MustCompile(`[a-zäöü]+`)

func CalculatePasswordStrength(p string) PasswordStrengthIndicator {
	var res PasswordStrengthIndicator
	res.MinLengthRequired = 8
	res.MaxLengthRequired = 1000

	// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#implement-proper-password-strength-controls
	strlen := utf8.RuneCount([]byte(p))
	if strlen >= res.MinLengthRequired {
		res.ContainsMinLength = true
	}

	// see https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#password-storage-cheat-sheet
	// and https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#compare-password-hashes-using-safe-functions
	if strlen <= res.MaxLengthRequired {
		// probably a DOS attack
		res.ContainsBelowMaxLength = true
	}

	if regexSpecialChar.MatchString(p) {
		res.ContainsSpecial = true
	}

	if regexNumber.MatchString(p) {
		res.ContainsNumber = true
	}

	if regexUppercase.MatchString(p) && regexLowercase.MatchString(p) {
		res.ContainsUpperAndLowercase = true
	}

	// see also https://reusablesec.blogspot.com/2010/10/new-paper-on-password-security-metrics.html
	// or https://www.omnicalculator.com/other/password-entropy
	if err := passwordvalidator.Validate(p, 60); err == nil {
		res.Complexity = Strong
		res.ComplexityScale = 0.8
	} else if err := passwordvalidator.Validate(p, 45); err == nil {
		res.Complexity = Weak
		res.ComplexityScale = 0.5
	} else {
		res.Complexity = VeryWeak
		res.ComplexityScale = 0.1
	}

	if strlen == 0 {
		res.ComplexityScale = 0
	}

	res.Acceptable = res.ContainsMinLength && res.ContainsSpecial && res.ContainsBelowMaxLength && res.ContainsNumber && res.ContainsUpperAndLowercase && res.Complexity > Weak
	if res.Acceptable {
		res.ComplexityScale = 1
	}

	return res
}
