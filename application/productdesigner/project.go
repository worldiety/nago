package productdesigner

type ProjectID string

type Project interface {
	Identity() ProjectID
}
