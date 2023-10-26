package serrors

// InfrastructureError is something like a broken I/O connection, disk full etc. which depends on the actual storage
// system. Even though these are all anticipated errors, the user cannot usually do something about it.
// Thus, the system mostly fails with an internal server error (500) at the presentation side.
// The responsibility to fix that is up to the service administrator.
type InfrastructureError interface {
	error
	isInfrastructure()
	Unwrap() error
}

type infrErr struct {
	Cause error
}

func (e infrErr) Unwrap() error {
	return e.Cause
}

func (e infrErr) isInfrastructure() {

}

func (e infrErr) Error() string {
	return "infrastructure error: " + e.Cause.Error()
}

// IntoInfrastructure returns an InfrastructureError if e is not nil. Otherwise, returns also nil.
func IntoInfrastructure(e error) InfrastructureError {
	if e == nil {
		return nil
	}

	return infrErr{Cause: e}
}
