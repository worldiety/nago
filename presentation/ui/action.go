package ui

type Action interface {
	isAction()
}

type Redirect struct {
	Target Target `json:"target"`
}

func (Redirect) isAction() {}

func (n Redirect) MarshalJSON() ([]byte, error) {
	return marshalJSON(n)
}

type Confirm struct {
	Action  Action
	Title   string
	Message string
	Confirm string
	Cancel  string
}

func (Confirm) isAction() {}
