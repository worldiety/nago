package ui2

import "net/http"

type Endpoint struct {
	Method  string // e.g. GET or POST
	Path    string
	Handler http.HandlerFunc
}
