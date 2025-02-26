package usercircle

import (
	"go.wdy.de/nago/application/user"
	"slices"
)

func NewIsCircleAdmin(repo Repository) IsCircleAdmin {
	return func(uid user.ID) (bool, error) {
		for circle, err := range repo.All() {
			if err != nil {
				return false, err
			}

			if slices.Contains(circle.Administrators, uid) {
				return true, nil
			}
		}

		return false, nil
	}
}
