package ui2

import "net/http"

type Context struct {
	writer  http.ResponseWriter
	request *http.Request
}

func newContext(w http.ResponseWriter, r *http.Request) Context {
	return Context{
		writer:  w,
		request: r,
	}
}

func Render(id PageID, scaffold Scaffold) {}

type response[T any] struct {
	Data T `json:"data"`
}

type PageID string

type NavItem struct {
	Title  string
	Action Navigation
}

func (n NavItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(n)
}

type Navigation struct {
	Target  PageID
	Payload any // optional arbitrary struct serialized e.g. into URL like identity or even form data?
}

func (n Navigation) MarshalJSON() ([]byte, error) {
	return marshalJSON(n)
}

type Persona interface {
	isPersona()
	Endpoints(page PageID, authenticated bool) []Endpoint
}
