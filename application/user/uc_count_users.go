package user

func NewCountUsers(repo Repository) CountUsers {
	return func() (int, error) {
		n, err := repo.Count()
		if err != nil {
			return 0, err
		}

		return n, nil
	}
}
