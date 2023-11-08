package ui

// Image represents something vector or bitmap image like. It could be a svg, a jpg or png or a font icon.
type Image interface {
	isView()
	isImage()
}
