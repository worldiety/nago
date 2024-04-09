package protocol

type Path string

type NewComponentRequested struct {
	Type      EventType         `json:"type" value:"NewComponentRequested"`
	Locale    string            `json:"activeLocale"`
	Path      Path              `json:"path" description:"This is the unique address for a specific component factory, e.g. my/component/path. This is typically a page."`
	Values    map[string]string `json:"values" description:"Contains string encoded parameters for a component, typically an entire page."`
	RequestId RequestId         `json:"requestId" description:"Request ID used to generate a new component request and is returned in the according response."`
}

type ComponentInvalidated struct {
	Type      EventType `json:"type" value:"ComponentInvalidated"`
	RequestId RequestId `json:"requestId" description:"Request ID from the NewComponentRequested event."`
	Component Component `json:"value" description:"The rendered component tree."`
}
