package ui

import (
	"encoding/json"
)

type redirectable interface {
	Redirection() Redirect
}

// Redirect offers the possibility to send non-http redirection requests (JSON) to the javascript shim,
// which evaluates the redirection response and applies the according navigation stack operation.
// This kind of navigation is not possible by using standard http redirects.
type Redirect struct {
	url       string
	direction string
	redirect  bool
}

func (r Redirect) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      string `json:"type"`
		URL       string `json:"url"`
		Direction string `json:"direction"`
		Redirect  bool   `json:"redirect"`
	}{
		Type:      "redirect",
		URL:       r.url,
		Direction: r.direction,
		Redirect:  r.redirect,
	})
}

func (r Redirect) Redirection() Redirect {
	return r
}

// Forward creates a new forward [Redirect] to the given url.
// Use query parameters to provide some (bookmarkable) contextual information.
func Forward(url string) Redirect {
	return Redirect{url: url, direction: "forward", redirect: true}
}
