package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type WebView struct {
	id         CID
	value      String
	properties []core.Property
}

func NewWebView(with func(*WebView)) *WebView {
	c := &WebView{
		id: nextPtr(),
	}

	c.value = NewShared[string]("value")

	c.properties = []core.Property{c.value}
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

func (c *WebView) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *WebView) Render() ora.Component {
	return c.render()
}

func (c *WebView) render() ora.WebView {
	return ora.WebView{
		Ptr:   c.id,
		Type:  ora.WebViewT,
		Value: c.value.render(),
	}
}
