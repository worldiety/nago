package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Shadow struct {
	Color  Color  `json:"c,omitempty"`
	Radius Length `json:"r,omitempty"`
	X      Length `json:"x,omitempty"`
	Y      Length `json:"y,omitempty"`
}

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

func (b Border) Radius(radius Length) Border {
	b.TopLeftRadius = radius
	b.TopRightRadius = radius
	b.BottomLeftRadius = radius
	b.BottomRightRadius = radius
	return b
}

func (b Border) Circle() Border {
	return b.Radius("999999dp")
}

func (b Border) Width(width Length) Border {
	b.LeftWidth = width
	b.TopWidth = width
	b.RightWidth = width
	b.BottomWidth = width
	return b
}

func (b Border) Color(c Color) Border {
	b.LeftColor = c
	b.TopColor = c
	b.BottomColor = c
	b.RightColor = c
	return b
}

func (b Border) Shadow(radius Length) Border {
	b.BoxShadow.Radius = radius
	b.BoxShadow.Color = "#00000054"
	b.BoxShadow.X = ""
	b.BoxShadow.Y = ""
	return b
}
