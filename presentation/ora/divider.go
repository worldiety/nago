package ora

type Divider struct {
	Ptr  Ptr           `json:"id"`
	Type ComponentType `json:"type" value:"Divider"`
	component
}
