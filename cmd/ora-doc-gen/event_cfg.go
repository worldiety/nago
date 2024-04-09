package main

import "go.wdy.de/nago/presentation/protocol"

func wConfigurationRequested(d *Doc) {
	d.Printf("### configuration requested\n")
	d.Printf(`
A backend developer has potentially defined a lot of configuration details about the application.
For example, there may be a color theme, customized icons, image resources, an application name and the available set of navigations, launch intents or other meta information.
It is expected, that this only happens once during initialization of the frontend process.

`)

	d.PrintSpec("Specification for a transaction", protocol.ConfigurationRequested{})
	d.PrintJSON("Example encoding for a transaction component", protocol.ConfigurationRequested{
		Type:           protocol.NewConfigurationRequestedT,
		AcceptLanguage: "fr-FR",
	})

	d.PrintTypescriptIface("Example typescript interface stub", protocol.ConfigurationRequested{})
}

func wConfigurationDefined(d *Doc) {
	d.Printf("### configuration defined\n")
	d.Printf(`
According to the locale request, string and svg resources can be localized by the backend. The returned locale is the actually picked locale from the requested locale query string.

It looks quite obfuscated, however this minified version is intentional, because it may succeed each transaction call.
A frontend may request acknowledges for each event, e.g. while typing in a text field, so this premature optimization is likely a win.
`)

	d.PrintSpec("Specification for a configuration defined event", protocol.ConfigurationDefined{})
	d.PrintJSON("Example encoding for a configuration defined event", protocol.ConfigurationDefined{
		Type:             protocol.ConfigurationDefinedT,
		ApplicationName:  "My Application",
		AvailableLocales: []string{"en_US", "de_DE"},
		ActiveLocale:     "en",
		Themes: protocol.Themes{
			Dark:  protocol.DefaultTheme(),
			Light: protocol.DefaultTheme(),
		},
		Resources: protocol.Resources{
			SVG: map[protocol.SVGID]protocol.SVGSrc{
				1: nextSVGSrc(),
				2: nextSVGSrc(),
				3: nextSVGSrc(),
			},
		},
	})

	d.PrintTypescriptIface("Example typescript interface stub", protocol.ConfigurationDefined{})
}
