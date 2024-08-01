package ora

// ConfigurationRequested is issued by the frontend to get the applications general configuration.
// A backend developer has potentially defined a lot of configuration details about the application.
// For example, there may be a color theme, customized icons, image resources, an application name and the available set of navigations, launch intents or other meta information.
// It is expected, that this only happens once during initialization of the frontend process.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ConfigurationRequested struct {
	Type           EventType   `json:"type" value:"ConfigurationRequested"`
	AcceptLanguage string      `json:"acceptLanguage"`
	ColorScheme    ColorScheme `json:"colorScheme" description:"HSLColor scheme hint which the frontend has picked. This may reduce graphical glitches, if the backend creates images or webview resources for the frontend."`
	WindowInfo     WindowInfo  `json:"windowInfo"`
	RequestId      RequestId   `json:"r" `
	event
}

func (e ConfigurationRequested) ReqID() RequestId {
	return e.RequestId
}

// A ConfigurationDefined event is the response to a [ConfigurationRequested] event.
// According to the locale request, string and svg resources can be localized by the backend.
// The returned locale is the actually picked locale from the requested locale query string.
//
// It looks quite obfuscated, however this minified version is intentional, because it may succeed each transaction call.
// A frontend may request acknowledges for each event, e.g. while typing in a text field, so this premature optimization is likely a win.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ConfigurationDefined struct {
	Type               EventType `json:"type" value:"ConfigurationDefined"`
	ApplicationID      string    `json:"applicationID"`
	ApplicationName    string    `json:"applicationName"`
	ApplicationVersion string    `json:"applicationVersion"`
	AvailableLocales   []string  `json:"availableLocales"`
	AppIcon            URI       `json:"appIcon"`
	ActiveLocale       string    `json:"activeLocale"`
	Themes             Themes    `json:"themes"`
	RequestId          RequestId `json:"r"`
	event
}

func (e ConfigurationDefined) ReqID() RequestId {
	return e.RequestId
}
