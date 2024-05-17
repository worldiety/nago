package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"io/fs"
)

type FileField struct {
	id           ora.Ptr
	label        String
	value        String
	hintLeft     String
	hintRight    String
	error        String
	multiple     Bool
	maxBytes     Int
	disabled     Bool
	filter       String
	properties   []core.Property
	fileReceiver func(fsys fs.FS) error
}

func NewFileField(with func(fileField *FileField)) *FileField {
	c := &FileField{
		id:        nextPtr(),
		label:     NewShared[string]("label"),
		value:     NewShared[string]("value"),
		hintLeft:  NewShared[string]("hintLeft"),
		hintRight: NewShared[string]("hintRight"),
		error:     NewShared[string]("error"),
		disabled:  NewShared[bool]("disabled"),
		multiple:  NewShared[bool]("multiple"),
		maxBytes:  NewShared[int64]("maxBytes"),
		filter:    NewShared[string]("filter"),
	}
	c.maxBytes.Set(1024 * 1024 * 16)

	c.properties = []core.Property{c.label, c.value, c.hintLeft, c.hintRight, c.error, c.disabled, c.disabled, c.multiple, c.maxBytes, c.filter}

	if with != nil {
		with(c)
	}

	return c
}

func (c *FileField) OnFilesReceived(fsys fs.FS) error {
	if c.fileReceiver != nil {
		return c.fileReceiver(fsys)
	}

	return nil
}

func (c *FileField) SetFilesReceiver(receiverCallback func(fsys fs.FS) error) {
	c.fileReceiver = receiverCallback
}

func (c *FileField) ID() ora.Ptr {
	return c.id
}

func (c *FileField) Value() String {
	return c.value
}

func (c *FileField) Label() String {
	return c.label
}

func (c *FileField) HintLeft() String {
	return c.hintLeft
}

func (c *FileField) HintRight() String {
	return c.hintRight
}

func (c *FileField) Accept() String {
	return c.filter
}

func (c *FileField) Multiple() Bool {
	return c.multiple
}

func (c *FileField) MaxBytes() Int {
	return c.maxBytes
}

func (c *FileField) Error() String {
	return c.error
}

func (c *FileField) Disabled() Bool {
	return c.disabled
}

func (c *FileField) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *FileField) Render() ora.Component {
	return c.render()
}

func (c *FileField) render() ora.FileField {
	return ora.FileField{
		Ptr:       c.id,
		Type:      ora.FileFieldT,
		Label:     c.label.render(),
		HintLeft:  c.hintLeft.render(),
		HintRight: c.hintRight.render(),
		Error:     c.error.render(),
		Disabled:  c.disabled.render(),
		Filter:    c.filter.render(),
		Multiple:  c.Multiple().render(),
		MaxBytes:  c.maxBytes.render(),
	}
}
