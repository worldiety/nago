package session

func NewFindUserSessionByID(repository Repository) FindUserSessionByID {
	return func(id ID) UserSession {
		return &sessionImpl{id: id, repo: repository}
	}
}
