package productdesigner

type Plane struct {
	Bleed Rect
	Safe  Rect
}

type Rect struct {
	Left, Top, Width, Height int
}
