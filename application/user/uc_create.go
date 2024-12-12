package user

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/enum"
	"go.wdy.de/nago/pkg/std"
	"strings"
	"sync"
	"time"
)

func NewCreate(mutex *sync.Mutex, findByMail FindByMail, repo Repository) Create {
	return func(subject permission.Auditable, model ShortRegistrationUser) (User, error) {
		if err := subject.Audit(PermCreate); err != nil {
			return User{}, err
		}

		// this is really harsh and allows intentionally only to create one user per second
		mutex.Lock()
		defer mutex.Unlock()

		if model.Password != model.PasswordRepeated {
			return User{}, std.NewLocalizedError("Eingabebeschränkung", "Die Kennwörter stimmen nicht überein.")
		}

		mail := Email(strings.ToLower(string(model.Email)))
		if !mail.Valid() {
			return User{}, fmt.Errorf("invalid email: %v", model.Email)
		}

		if err := model.Password.Validate(); err != nil {
			return User{}, err
		}

		salt, hash, err := model.Password.Hash(Argon2IdMin)
		if err != nil {
			return User{}, err
		}

		createdAt := time.Now()
		user := User{
			// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#user-ids
			ID:        data.RandIdent[ID](),
			Email:     mail,
			Algorithm: Argon2IdMin,
			Contact: Contact{
				Firstname:         model.Firstname,
				Lastname:          model.Lastname,
				PreferredLanguage: model.PreferredLanguage.String(),
			},
			Salt:                  salt,
			PasswordHash:          hash,
			CreatedAt:             createdAt,
			LastPasswordChangedAt: createdAt,
			Status:                enum.Make[AccountStatus](Enabled{}),
		}

		// intentionally validate now, so that an attacker cannot use this method to massively
		// find out, which mails exist in the system
		optView, err := findByMail(subject, mail)
		if err != nil {
			return User{}, fmt.Errorf("cannot check for existing user: %w", err)
		}

		if optView.IsSome() {
			// actually this allows to find out, that a certain user is available. However, there is simply no
			// other possibility to not expose that information. We can only nag the attacker using excessive
			// slow-downs.
			time.Sleep(5 * time.Second)
			return User{}, std.NewLocalizedError("Nutzerregistrierung", "Die E-Mail-Adresse wird bereits verwendet.")
		}

		// unlikely, but better safe than sorry
		optUsr, err := repo.FindByID(user.ID)
		if err != nil {
			return User{}, fmt.Errorf("cannot find user by id: %w", err)
		}

		if optUsr.IsSome() {
			return User{}, fmt.Errorf("user id already taken")
		}

		// persist
		err = repo.Save(user)
		if err != nil {
			return User{}, fmt.Errorf("cannot persist new user: %w", err)
		}

		return user, nil
	}
}
