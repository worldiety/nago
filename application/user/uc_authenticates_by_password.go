package user

import (
	"go.wdy.de/nago/pkg/std"
)

func NewAuthenticatesByPassword(userByMail FindByMail, system System) AuthenticateByPassword {
	return func(email Email, password Password) (std.Option[User], error) {

		if !email.Valid() {
			return std.None[User](), std.NewLocalizedError("Login nicht möglich", "Dieses EMail-Adressformat ist nicht erlaubt.")
		}

		optUsr, err := userByMail(system(), email)
		if err != nil {
			return std.None[User](), err
		}

		// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#authentication-responses

		if optUsr.IsNone() {
			return std.None[User](), noLoginErr
		}

		usr := optUsr.Unwrap()
		if err := password.CompareHashAndPassword(usr.Algorithm, usr.Salt, usr.PasswordHash); err != nil {
			return std.None[User](), err
		}

		if !usr.Enabled() {
			return std.None[User](), noLoginErr
		}

		return optUsr, nil
	}
}
