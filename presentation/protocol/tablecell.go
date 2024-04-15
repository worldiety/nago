package protocol

type TableCell struct {
	Ptr  Ptr                 `json:"id"`
	Type ComponentType       `json:"type" value:"TableCell"`
	Body Property[Component] `json:"body"`

	component
}
