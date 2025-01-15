package user

import (
	"go.wdy.de/nago/pkg/std"
)

const EMailNotVerifiedErr std.Error = "email not verified"

func NewAuthenticatesByPassword(userByMail FindByMail, system SysUser) AuthenticateByPassword {
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

		if !usr.EMailVerified {
			// security note: intentionally it is not safe to let the user login, if his EMail was never
			// verified. This opens up all kinds of identity stealing by default, even though this may
			// be common in the world of shopping systems
			return std.None[User](), std.NewLocalizedError("Login nicht möglich", "Das Konto muss zuerst bestätigt werden").WithError(EMailNotVerifiedErr)
		}

		return optUsr, nil
	}
}
