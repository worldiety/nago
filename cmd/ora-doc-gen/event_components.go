package main

import "go.wdy.de/nago/presentation/protocol"

func wNewComponentRequested(d *Doc) {
	d.Printf("### new component requested\n")
	d.Printf(`
A frontend must allocate an addressable component explicitely in the backend within its channel scope.
Adressable components are like pages in a classic server side rendering or like routing targets in single page apps.
We do not call them _page_ anymore, because that has wrong assocations in the web world.
Adressable components exist independently from each other and share no lifecycle with each other.
However, a frontend can create as many component instances it wants.
It does not matter, if these components are of the same type, addresses or entirely different.
The backend responds with a component invalidation event.

Factories of addressable components are always stateless.
However, often it does not make sense without additional parameters, e.g. because a detail view needs to know which entity has to be displayed.
`)

	d.PrintSpec("Specification for a new component requested event", protocol.NewComponentRequested{})
	d.PrintJSON("Example encoding for a new component requested event", protocol.NewComponentRequested{
		Type:   protocol.NewComponentRequestedT,
		Locale: "de_DE",
		Path:   "invoices/details/invoice",
		Values: map[string]string{
			"invoice": nextUUID(),
			"tabIdx":  "3",
		},
		RequestId: nextRequestId(),
	})

	d.PrintTypescriptIface("Example typescript interface stub", protocol.ComponentInvalidated{})
}

func wComponentInvalidated(d *Doc) {
	d.Printf("### component invalidated\n")

	d.PrintSpec("Specification for a transaction", protocol.ComponentInvalidated{})
	d.PrintJSON("Example encoding for a transaction component", protocol.ComponentInvalidated{
		Type:      protocol.ComponentInvalidatedT,
		Component: newButton(),
		RequestId: nextRequestId(),
	})
}
