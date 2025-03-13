package usercircle

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"slices"
)

func NewFindByID(repo Repository) FindByID {
	return func(subject auth.Subject, id ID) (std.Option[Circle], error) {
		if !subject.Valid() {
			return std.None[Circle](), user.InvalidSubjectErr
		}
		
		optCircle, err := repo.FindByID(id)
		if err != nil {
			return std.None[Circle](), err
		}

		if optCircle.IsNone() {
			return std.None[Circle](), nil
		}

		circle := optCircle.Unwrap()
		if slices.Contains(circle.Administrators, subject.ID()) {
			return optCircle, nil
		}

		if err := subject.Audit(PermFindByID); err != nil {
			return std.None[Circle](), err
		}

		return optCircle, nil
	}
}
