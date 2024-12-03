package application

import "go.wdy.de/nago/auth/iam"

type UserManagement struct {
	ChangeMyPassword func() iam.ChangeMyPassword
}
