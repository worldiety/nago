package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Shadow struct {
	Color  Color  `json:"c,omitempty"`
	Radius Length `json:"r,omitempty"`
	X      Length `json:"x,omitempty"`
	Y      Length `json:"y,omitempty"`
}

// Border adds the defined border and dimension to the component. Note, that a border will change the dimension.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Border struct {
	TopLeftRadius     Length `json:"tlr,omitempty"`
	TopRightRadius    Length `json:"trr,omitempty"`
	BottomLeftRadius  Length `json:"blr,omitempty"`
	BottomRightRadius Length `json:"brr,omitempty"`

	LeftWidth   Length `json:"lw,omitempty"`
	TopWidth    Length `json:"tw,omitempty"`
	RightWidth  Length `json:"rw,omitempty"`
	BottomWidth Length `json:"bw,omitempty"`

	LeftColor   Color `json:"lc,omitempty"`
	TopColor    Color `json:"tc,omitempty"`
	RightColor  Color `json:"rc,omitempty"`
	BottomColor Color `json:"bc,omitempty"`

	BoxShadow Shadow `json:"s,omitempty"`
}
