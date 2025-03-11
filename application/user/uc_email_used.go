package user

func NewEMailUsed(repo Repository) EMailUsed {
	return func(email Email) (bool, error) {
		for user, err := range repo.All() {
			if err != nil {
				return false, err
			}

			if user.Email.Equals(email) {
				return true, nil
			}
		}

		return false, nil
	}
}
