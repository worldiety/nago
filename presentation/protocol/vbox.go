package protocol

type VBox struct {
	Ptr      Ptr                   `json:"id"`
	Type     ComponentType         `json:"type" value:"VBox"`
	Children Property[[]Component] `json:"children"`
	component
}
