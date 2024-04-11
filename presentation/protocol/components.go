package protocol

// A ComponentFactoryId identifies a unique constructor for a specific ComponentType.
// Such an addressable Component is likely a page and instantiated and rendered.
// In return, a ComponentInvalidated event will be sent in the future.
// For details, see the [NewComponentRequested] event.
type ComponentFactoryId string

// NewComponentRequested allocates an addressable component explicitely in the backend within its channel scope.
// Adressable components are like pages in a classic server side rendering or like routing targets in single page apps.
// We do not call them _page_ anymore, because that has wrong assocations in the web world.
// Adressable components exist independently from each other and share no lifecycle with each other.
// However, a frontend can create as many component instances it wants.
// It does not matter, if these components are of the same type, addresses or entirely different.
// The backend responds with a component invalidation event.
//
// Factories of addressable components are always stateless.
// However, often it does not make sense without additional parameters, e.g. because a detail view needs to know which entity has to be displayed.
type NewComponentRequested struct {
	Type      EventType          `json:"type" value:"NewComponentRequested"`
	Locale    string             `json:"activeLocale" description:"This locale has been picked by the backend."`
	Factory   ComponentFactoryId `json:"factory" description:"This is the unique address for a specific component factory, e.g. my/component/path. This is typically a page."`
	Values    map[string]string  `json:"values" description:"Contains string encoded parameters for a component. This is like query parameters."`
	RequestId RequestId          `json:"requestId" description:"Request ID used to generate a new component request and is returned in the according response."`
	_         struct{}           `NewComponentRequested allocates an addressable component explicitely in the backend within its channel scope.
Adressable components are like pages in a classic server side rendering or like routing targets in single page apps.
We do not call them _page_ anymore, because that has wrong assocations in the web world.
Adressable components exist independently from each other and share no lifecycle with each other.
However, a frontend can create as many component instances it wants.
It does not matter, if these components are of the same type, addresses or entirely different.
The backend responds with a component invalidation event.

Factories of addressable components are always stateless.
However, often it does not make sense without additional parameters, e.g. because a detail view needs to know which entity has to be displayed.
`
	event
}

type ComponentInvalidationRequested struct {
	Type      EventType `json:"type" value:"ComponentInvalidationRequested"`
	RequestId RequestId `json:"requestId" description:"Request ID from the NewComponentRequested event."`
	Component Ptr       `json:"ptr" description:"The pointer of the component, which shall be rendered again. Only Pointer created with NewComponentRequested are valid."`
	event
}

type ComponentDestructionRequested struct {
	Type      EventType `json:"type" value:"ComponentDestructionRequested"`
	RequestId RequestId `json:"requestId" description:"Request ID."`
	Component Ptr       `json:"ptr" description:"The pointer of the component, which shall be rendered again. Only Pointer created with NewComponentRequested are valid."`
	event
}

type ComponentInvalidated struct {
	Type      EventType `json:"type" value:"ComponentInvalidated"`
	RequestId RequestId `json:"requestId" description:"Request ID from the ComponentInvalidationRequested or NewComponentRequested event."`
	Component Component `json:"value" description:"The rendered component tree."`

	event
}

type ErrorOccurred struct {
	Type      EventType `json:"type" value:"ErrorOccurred"`
	RequestId RequestId `json:"requestId" description:"Request ID from the NewComponentRequested event."`
	Message   string    `json:"message" description:"A message describing the error."`
	event
}
