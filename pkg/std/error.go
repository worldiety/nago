package std

// LocalizedError provides a for the current context already translated error message.
// See [NewLocalizedError].
type LocalizedError interface {
	error
	Unwrap() error
	Title() string
	Description() string
}

type localizedError struct {
	title string
	desc  string
	cause error
}

func (e localizedError) Error() string {
	return e.desc
}

func (e localizedError) Title() string {
	return e.title
}

func (e localizedError) Description() string {
	return e.desc
}

func (e localizedError) Unwrap() error {
	return e.cause
}

func NewLocalizedError(title, desc string) LocalizedError {
	return localizedError{
		title: title,
		desc:  desc,
	}
}
