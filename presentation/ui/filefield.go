package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"io"
)

type FileUpload interface {
	Size() int64
	Name() string
	Open() (io.ReadSeekCloser, error)
	Sys() any
}

type UploadToken string

type FileHandler func(files []FileUpload)

type FileField struct {
	id               ora.Ptr
	label            String
	value            String
	hintLeft         String
	hintRight        String
	error            String
	multiple         Bool
	maxBytes         Int
	disabled         Bool
	filter           String
	uploadToken      String
	onUploadReceived FileHandler
	properties       []core.Property
}

func NewFileField(with func(fileField *FileField)) *FileField {
	c := &FileField{
		id:          nextPtr(),
		label:       NewShared[string]("label"),
		value:       NewShared[string]("value"),
		hintLeft:    NewShared[string]("hintLeft"),
		hintRight:   NewShared[string]("hintRight"),
		error:       NewShared[string]("error"),
		disabled:    NewShared[bool]("disabled"),
		multiple:    NewShared[bool]("multiple"),
		maxBytes:    NewShared[int64]("maxBytes"),
		filter:      NewShared[string]("filter"),
		uploadToken: NewShared[string]("uploadToken"),
	}

	c.uploadToken.Set(nextToken())

	c.properties = []core.Property{c.label, c.value, c.hintLeft, c.hintRight, c.error, c.disabled, c.disabled, c.multiple, c.maxBytes, c.filter, c.uploadToken}

	if with != nil {
		with(c)
	}

	return c
}

// OnUploadReceived is not a property, because we have special quirks in the background
// to process the upload outside of our message queue. The given files are only valid during the
// call of the handler, thus it is invalid to keep the given FileUpload or open streams. You
// must consume the file within the handler, read your data and return.
// TODO unify message handling and make it single threaded for all kind of events: ui, inter-domain, intra-domain, uploads and rest-calls !?!
func (c *FileField) OnUploadReceived(f FileHandler) {
	c.onUploadReceived = f
	c.hintLeft.SetDirty(true)  // fake re-render, today we always re-render anyway
	c.hintRight.SetDirty(true) // fake re-render, today we always re-render anyway
}

func (c *FileField) UploadToken() UploadToken {
	return UploadToken(c.uploadToken.Get())
}

func (c *FileField) getOnUploadReceived() FileHandler {
	return c.onUploadReceived
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
		Ptr:         c.id,
		Type:        ora.FileFieldT,
		Label:       c.label.render(),
		HintLeft:    c.hintLeft.render(),
		HintRight:   c.hintRight.render(),
		Error:       c.error.render(),
		Disabled:    c.disabled.render(),
		Filter:      c.filter.render(),
		Multiple:    c.Multiple().render(),
		MaxBytes:    c.maxBytes.render(),
		UploadToken: c.uploadToken.render(),
	}
}
