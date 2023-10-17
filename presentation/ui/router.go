package ui

type Route struct {
	Pattern string
	Render  func(event RouteEvent) Page
}

type RouteEvent struct {
}
