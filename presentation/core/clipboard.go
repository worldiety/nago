package core

import (
	"fmt"
	"go.wdy.de/nago/presentation/proto"
)

type Clipboard interface {
	// SetText writes plain text data into the clipboard.
	SetText(text string) error

	// Text reads plain text data from the clipboard. The callback is invoked, when the clipboard is done.
	// Due to distributed systems nature, this may never return.
	Text(onResult func(text string, err error))
}

type clipboardController struct {
	wnd *scopeWindow
}

func (c *clipboardController) SetText(text string) error {
	c.wnd.parent.Publish(&proto.ClipboardWriteTextRequested{
		Text: proto.Str(text),
	})
	
	return nil
}

func (c *clipboardController) Text(onResult func(text string, err error)) {
	onResult("", fmt.Errorf("clipboard read not implemented"))
}

func newClipboardController(wnd *scopeWindow) *clipboardController {
	return &clipboardController{wnd: wnd}
}
