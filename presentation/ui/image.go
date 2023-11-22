package ui

type Image interface {
	isImage()
}

// spec only
type image struct{}

func (image) isImage() {}

func (image) JSONSchemaAnyOf() []interface{} {
	return []any{FontIcon{}, ImageURL{}}
}

// FontIcon see also https://fonts.google.com/icons, prefixed by mdi- e.g. like "mdi-home".
type FontIcon struct {
	Name  string `json:"name"`
	Color string `json:"color"` // only a hint and may be ignored, e.g. primary, error or a color code?
}

func (n FontIcon) isImage() {}

func (n FontIcon) MarshalJSON() ([]byte, error) {
	return marshalJSON(n)
}

type ImageURL struct {
	URL string `json:"URL"`
}

func (n ImageURL) isImage() {}
