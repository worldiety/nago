package ora

type Card struct {
	Ptr      Ptr                   `json:"id"`
	Type     ComponentType         `json:"type" value:"Card"`
	Children Property[[]Component] `json:"children"`
	Action   Property[Ptr]         `json:"action"`
	component
}
