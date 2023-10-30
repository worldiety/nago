package ui

type NavigationRequest interface {
	isNavigationRequest()
}

type PageID string

type Navbar struct {
	Caption string
}
