package session

const (
	NotLoggedInErr NotLoggedInError = "not logged in"
)

type NotLoggedInError string

func (e NotLoggedInError) Error() string {
	return string(e)
}

func (e NotLoggedInError) PermissionDenied() bool {
	return true
}

func (e NotLoggedInError) NotLoggedIn() bool {
	return true
}
