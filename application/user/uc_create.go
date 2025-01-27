package user

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
	"golang.org/x/text/language"
	"log/slog"
	"strings"
	"sync"
	"time"
)

func NewCreate(mutex *sync.Mutex, eventBus events.EventBus, findByMail FindByMail, repo Repository) Create {
	return func(subject permission.Auditable, model ShortRegistrationUser) (User, error) {
		if err := subject.Audit(PermCreate); err != nil {
			return User{}, err
		}

		// this is really harsh and allows intentionally only to create one user per second
		mutex.Lock()
		defer mutex.Unlock()

		requiredPasswordChange := false
		if model.Password == "" && model.PasswordRepeated == "" {
			model.Password = data.RandIdent[Password]()
			model.PasswordRepeated = model.Password
			requiredPasswordChange = true
		}

		if model.Password != model.PasswordRepeated {
			return User{}, std.NewLocalizedError("Eingabebeschränkung", "Die Kennwörter stimmen nicht überein.")
		}

		mail := Email(strings.ToLower(string(model.Email)))
		if !mail.Valid() {
			return User{}, std.NewLocalizedError("Eingabebeschränkung", "Auch wenn es sich um eine potentiell gültige E-Mail Adresse handeln könnte, wird dieses Format nicht unterstützt.")
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
				Firstname:       model.Firstname,
				Lastname:        model.Lastname,
				DisplayLanguage: model.PreferredLanguage.String(),
			},
			Salt:                  salt,
			PasswordHash:          hash,
			CreatedAt:             createdAt,
			LastPasswordChangedAt: createdAt,
			Status:                Enabled{},
			EMailVerified:         model.Verified,
			RequirePasswordChange: requiredPasswordChange,
			// initially, give the user a week to respond. Note, that for self registration we just may
			// remove users which have never been verified automatically
			VerificationCode: NewCode(DefaultVerificationLifeTime),
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

		if model.NotifyUser {
			tag, err := language.Parse(user.Contact.DisplayLanguage)
			if err != nil {
				slog.Error("user contact has invalid preferred language", "err", err)
			}

			eventBus.Publish(Created{
				ID:                user.ID,
				Firstname:         user.Contact.Firstname,
				Lastname:          user.Contact.Lastname,
				Email:             user.Email,
				PreferredLanguage: tag,
				NotifyUser:        model.NotifyUser,
				VerificationCode:  user.VerificationCode,
			})
		}

		return user, nil
	}
}
