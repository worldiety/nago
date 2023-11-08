package data

// TODO this smells a lot like UI/presentation
type Iconic interface {
	isIcon()
}

// TODO is this a field???
type Validateable[T any] struct {
	Value          T
	ErrorText      string // clearly a domain description
	SupportingText string // not sure if purely presentation or also context dependent => domain driven
	LabelText      string // probably solely presentation
	Icon           Iconic // probably solely presentation
	ReadOnly       bool   // not sure, if this occurs at the domain layer, it is probably weekly modelled, because there must be a read model and a write model, however how should a "stupid" UI know the transformation?
}

// TODO this must be treated special
func (v Validateable[T]) MarshalJSONRepository() ([]byte, error) {
	panic("todo")
}

func (v Validateable[T]) IsReadOnly() bool {
	return v.ReadOnly
}

func (v Validateable[T]) LeadingIcon() Iconic {
	return v.Icon
}

func (v Validateable[T]) Error() string {
	return v.ErrorText
}

func (v Validateable[T]) Supporting() string {
	return v.SupportingText
}

func (v Validateable[T]) Label() string {
	return v.LabelText
}
