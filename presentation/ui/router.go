package ui

type Route struct {
	Pattern string
	Render  func(event RouteEvent) View
}

type RouteEvent struct {
}
