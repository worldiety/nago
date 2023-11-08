package ui2

type Status int

func (s Status) Ok() bool {
	return s == Ok
}

const (
	Ok Status = iota
	Unauthorized
	NoLogin
	NotFound
	InternalServerError // TODO we must display the incident code
)
