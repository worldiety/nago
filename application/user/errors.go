package user

const (
	InvalidSubjectErr   InvalidSubjectError   = "invalid subject"
	PermissionDeniedErr PermissionDeniedError = "permission denied"
)

type InvalidSubjectError string

func (e InvalidSubjectError) Error() string {
	return string(e)
}

func (e InvalidSubjectError) PermissionDenied() bool {
	return true
}

func (e InvalidSubjectError) NotLoggedIn() bool {
	return true
}

type PermissionDeniedError string

func (e PermissionDeniedError) Error() string {
	return string(e)
}

func (e PermissionDeniedError) PermissionDenied() bool {
	return true
}
