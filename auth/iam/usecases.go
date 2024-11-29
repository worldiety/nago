package iam

import "go.wdy.de/nago/auth"

type ChangeMyPassword func(subject auth.Subject, oldPassword, newPassword, newRepeated Password) error

func NewChangeMyPassword(service *Service) ChangeMyPassword {
	return func(subject auth.Subject, oldPassword, newPassword, newRepeated Password) error {
		_, err := service.ChangeMyPassword(subject, oldPassword, newPassword, newRepeated)
		return err
	}
}
