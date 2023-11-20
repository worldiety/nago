package ui

import (
	"net/http"
)

type Void struct{}

type FormAction[Form, PageParams any] struct {
	Title   string
	Receive func(Form, PageParams) (Form, Action) // returns either F on error or the action to perform
}

type EventHandler[E any] struct {
	Action    E       // the thing to send
	OnReceive func(E) // received it
}

type Context struct {
	writer  http.ResponseWriter
	request *http.Request
}

type response[T any] struct {
	Data T `json:"data"`
}

type NavItem struct {
	Title  string
	Target Target
	Icon   FontIcon
}

func (n NavItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(n)
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
