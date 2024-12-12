package user

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
)

func NewFindByMail(repository Repository) FindByMail {
	return func(subject permission.Auditable, email Email) (std.Option[User], error) {
		if err := subject.Audit(PermFindByMail); err != nil {
			return std.None[User](), err
		}

		// TODO this is really slow O(n), we either need some cache or an inverse index
		for user, err := range repository.All() {
			if err != nil {
				return std.None[User](), fmt.Errorf("cannot loop user repo: %w", err)
			}

			if user.Email == email {
				return std.Some(user), nil
			}
		}

		return std.None[User](), nil
	}
}
