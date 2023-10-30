package ui

type NavigationRequest interface {
	isNavigationRequest()
}

type PageID string

type Navbar struct {
	Caption string
}

type Forward struct {
	Target PageID
}

type Backward struct {
	Target PageID
}
