package ui2

import (
	"go.wdy.de/nago/container/data"
	"net/http"
)

type PersonenForm struct {
	ID   ComponentID
	Name data.Validateable[string]
}

type GoTo struct {
	Target ComponentID
	Title  string // optional?
	Params any    //??? we cannot prove statically the connection a another component
}

func (GoTo) isAction() {}

type Confirm struct {
	Action  Action
	Title   string
	Message string
	Confirm string
	Cancel  string
}

func (Confirm) isAction() {}

type Void struct{}

type InputText struct{}

func (InputText) isFieldOrRow() {}
func (InputText) isField()      {}

// e.g. /api/v1/page/hello-world/{super-id}/users/list?{sort-order}
// what about SSR (htmx?)
// e.g. /hello-world/{super-id} => partials using "accept-content"???
type ExampleParams struct {
	SuperID   string `path:"super-id"`
	SortOrder string `query:"sort-order"`
}

// TODO params just per page because component-api is already page-scoped?
type Form[FormType, Params any] struct {
	ID     ComponentID
	Submit FormAction[FormType, Params]
	Load   func(Params) FormType
}

type FormAction[Form, Params any] struct {
	Title    string
	OnSubmit func(Form) (Form, Action) // returns either F on error or the action to perform
}

type MyForm struct {
	PersonalInfo `caption:"Persönliche Informationen"`
	AdressInfo   `caption:"Adresse"`
}

type PersonalInfo struct {
	Firstname InputText
	Lastname  InputText
}

type AdressInfo struct {
	Zip        InputText
	FileUpload []byte
}

type FormSection struct {
	Caption  string
	Children []FieldOrRow
}

type FieldOrRow interface {
	isFieldOrRow()
}

type Field interface {
	isField()
	isFieldOrRow()
}

type FormRow struct {
	Children []Field
}

func (FormRow) isFieldOrRow() {

}

type EventHandler[E any] struct {
	Action    E       // the thing to send
	OnReceive func(E) // received it
}

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

type NavItem struct {
	Title  string
	Action Navigation
	Icon   FontIcon
}

func (n NavItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(n)
}

type Navigation struct {
	Target  PageID
	Payload any // optional arbitrary struct serialized e.g. into URL like identity or even form data?
}

type Action interface {
	isAction()
}

func (n Navigation) MarshalJSON() ([]byte, error) {
	return marshalJSON(n)
}

type Persona interface {
	Id() ComponentID
	Endpoints(page PageID, authenticated bool) []Endpoint
}

func must2[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
