package main

import "go.wdy.de/nago/presentation/protocol"

func wButton(d *Doc) {
	d.Printf("### Button\n\n")
	d.PrintSpec("Specification for a button component", protocol.Button{})
	d.PrintJSON("Example encoding for a button component", newButton())

	d.PrintTypescriptIface("Example typescript interface stub", protocol.Button{})

}

func newButton() protocol.Button {
	return protocol.Button{
		Ptr:  nextPtr(),
		Type: protocol.ButtonT,
		Caption: protocol.Property[string]{
			Ptr:   nextPtr(),
			Value: nextCaption(),
		},
		PreIcon: protocol.Property[protocol.SVGSrc]{
			Ptr:   nextPtr(),
			Value: nextSVGSrc(),
		},
		PostIcon: protocol.Property[protocol.SVGSrc]{
			Ptr:   nextPtr(),
			Value: nextSVGSrc(),
		},
		Color:    protocol.Property[protocol.Color]{},
		Disabled: protocol.Property[bool]{},
		Action:   protocol.Property[protocol.Ptr]{},
	}
}
