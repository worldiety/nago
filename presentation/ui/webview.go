package ui

import "go.wdy.de/nago/container/slice"

type WebView struct {
	id         CID
	value      String
	properties slice.Slice[Property]
}

func NewWebView(with func(*WebView)) *WebView {
	c := &WebView{
		id: nextPtr(),
	}

	c.value = NewShared[string]("value")

	c.properties = slice.Of[Property](c.value)
	if with != nil {
		with(c)
	}

	return c
}

// Value provides access to raw HTML. This is prone to any kind of injection attacks by definition, so ensure
// that your html comes from a trusted source, e.g. embedded or generated html/template etc. It is not guaranteed
// that a full-featured webbrowser with all capabilities will interpret this.
func (c *WebView) Value() String {
	return c.value
}

func (c *WebView) ID() CID {
	return c.id
}

func (c *WebView) Type() string {
	return "WebView"
}

func (c *WebView) Properties() slice.Slice[Property] {
	return c.properties
}
