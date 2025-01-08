package secret

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"slices"
)

func NewFindMySecretByID(repository Repository) FindMySecretByID {
	return func(subject auth.Subject, id ID) (std.Option[Secret], error) {
		if err := subject.Audit(PermFindMySecrets); err != nil {
			return std.Option[Secret]{}, err
		}

		optSecret, err := repository.FindByID(id)
		if err != nil {
			return std.Option[Secret]{}, err
		}

		if optSecret.IsNone() {
			return optSecret, nil
		}

		src := optSecret.Unwrap()
		if slices.Contains(src.Owners, subject.ID()) {
			return optSecret, nil
		}

		return std.Option[Secret]{}, std.NewLocalizedError("Zugriff verweigert", "Nur Besitzer des Secrets dürfen das Geheimnis einsehen.")
	}
}
