package usercircle

import (
	"go.wdy.de/nago/application/user"
)

func NewIsMyCircleMember(repo Repository, findUserByID user.FindByID) IsMyCircleMember {
	return func(uid user.ID, cid ID, other user.ID) (bool, error) {
		circle, err := myCircle(repo, uid, cid)
		if err != nil {
			return false, err
		}

		optUsr, err := findUserByID(user.SU(), other)
		if err != nil {
			return false, err
		}

		if optUsr.IsNone() {
			return false, nil
		}
		
		usr := optUsr.Unwrap()
		return circle.isMember(usr), nil

	}
}
