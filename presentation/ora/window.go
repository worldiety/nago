package ora

// DP is Density-independent pixels: an abstract unit that is based on the physical density of the screen.
// These units are relative to a 160 dpi (dots per inch) screen, on which 1 dp is roughly equal to 1 px.
// When running on a higher density screen, the number of pixels used to draw 1 dp is scaled up by a factor
// appropriate for the screen's dpi.
//
// Likewise, when on a lower-density screen, the number of pixels used for 1 dp is scaled down.
// The ratio of dps to pixels changes with the screen density, but not necessarily in direct proportion.
// Using dp units instead of px units is a solution to making the view dimensions in your layout
// resize properly for different screen densities. It provides consistency for the real-world sizes of
// your UI elements across different devices.
// Source: https://developer.android.com/guide/topics/resources/more-resources.html#Dimension
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type DP float64

// Density describes the scale factor of physical pixels to screen pixels normalized to a 160dpi screen.
// This is identical to the Android specification. On a 160dpi screen, this factor is 1. Note, that
// this may also be used to optimize accessibility which makes everything equally larger. There is also the
// concept of SP, but that is usually implemented at the frontend interpreter anyway.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Density float64

// WindowSizeClass represents media break points of the screen which an ora application is shown.
// The definition of a size class is disjunct and for all possible sizes, exact one size class will match.
// See also https://developer.android.com/develop/ui/views/layout/window-size-classes.
type WindowSizeClass struct {
	// this looks weired, however it makes the WindowSizeClass comparable, because we compare
	// the box pointer instead of the undefined func pointer
	box *struct {
		check func(width, height DP) bool
	}
	name string
}

func (w WindowSizeClass) Match(width, height DP) bool {
	return w.box.check(width, height)
}

func (w WindowSizeClass) String() string {
	return w.name
}

var ExpandedWindow = WindowSizeClass{
	name: "Expanded",
	box: &struct {
		check func(width DP, height DP) bool
	}{check: func(width DP, height DP) bool {
		return width >= 840
	}},
}

var MediumWindow = WindowSizeClass{
	name: "Medium",
	box: &struct {
		check func(width DP, height DP) bool
	}{check: func(width DP, height DP) bool {
		return 600 <= width && width < 840
	}},
}
var CompactWindow = WindowSizeClass{
	name: "Compact",
	box: &struct {
		check func(width DP, height DP) bool
	}{check: func(width DP, height DP) bool {
		return width < 600
	}},
}

// ComputeSizeClass takes the given width and height and returns exactly one of the Ora specified size classes.
// You can simply compare a size class with others.
func ComputeSizeClass(width, height DP) WindowSizeClass {
	switch {
	case ExpandedWindow.Match(width, height):
		return ExpandedWindow
	case MediumWindow.Match(width, height):
		return MediumWindow
	case CompactWindow.Match(width, height):
		return CompactWindow
	default:
		panic("unreachable")
	}
}

// WindowInfo describes the area into which the frontend renders the ora view tree.
// A user can simply change the layout of the screen, e.g. by rotation the smartphone or
// changing the size of a browser window.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type WindowInfo struct {
	Width   DP      `json:"width"`
	Height  DP      `json:"height"`
	Density Density `json:"density"`
}

func (w WindowInfo) SizeClass() WindowSizeClass {
	return ComputeSizeClass(w.Width, w.Height)
}

// WindowInfoChanged is raised by the frontend whenever the window metrics changed in a significant way.
// It is not guaranteed that every pixel change will trigger such an event.
// However, a frontend must guarantee to send such an event if the WindowSizeClass is changed.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type WindowInfoChanged struct {
	Type EventType  `json:"type" value:"WindowInfoChanged"`
	Info WindowInfo `json:"info"`

	event
}

func (e WindowInfoChanged) ReqID() RequestId {
	return 0
}
