package main

import "go.wdy.de/nago/presentation/protocol"

func wTx(d *Doc) {
	d.Printf("### aggregated events\n")
	d.Printf(`
A transaction forms an envelope message which contains a bunch of the actual events, which shall be applied within a single event processing step at the receivers side in exactly the given order.
A receiver must ensure the sequential processing of the contained messages and must not apply them in different order, partially or in parallel. Nested transactions are invalid.

It looks quite obfuscated, however this minified version is intentional.
For example, a frontend may issue aggregated events for each keystroke (setting a property and calling a func) so this premature optimization is likely a win.


The _requestId_ is optional and its content is an arbitrary value from the sender.
If the _requestId_ is neither null nor empty, the receiver must respond with an _ack_ event.
The _ack_ event must be the first message in the next transaction from the receiver.
However, due to channel interruptions, the _ack_ may get lost, thus a participant must handle this gracefully using a timeout mechanism.
The frontend must not freeze, but shall instead visualize the waiting, e.g. by debouncing interactive elements or by even disabling the entire screen and showing an indeterminate progress.

`)

	d.PrintSpec("Specification for a transaction", protocol.EventsAggregated{})
	d.PrintJSON("Example encoding for a transaction component", protocol.EventsAggregated{
		Type: protocol.EventsAggregatedT,
		Events: []protocol.Event{
			protocol.SetPropertyValueRequested{
				Type:  protocol.SetPropertyValueRequestedT,
				Ptr:   nextPtr(),
				Value: "hello world",
			},
			protocol.FunctionCallRequested{
				Type: protocol.FunctionCallRequestedT,
				Ptr:  nextPtr(),
			},
		},
		RequestId: nextRequestId(),
	})

	d.PrintTypescriptIface("Example typescript interface stub", protocol.EventsAggregated{})
}

func wAck(d *Doc) {
	d.Printf("### acknowledged event\n")
	d.Printf(`
A transaction forms an envelope message which contains a bunch of the actual events, which shall be applied within a single event processing step at the receivers side in exactly the given order.
A receiver must ensure the sequential processing of the contained messages and must not apply them in different order, partially or in parallel. Nested transactions are invalid.

It looks quite obfuscated, however this minified version is intentional, because it may succeed each transaction call.
A frontend may request acknowledges for each event, e.g. while typing in a text field, so this premature optimization is likely a win.
`)

	d.PrintSpec("Specification for an acknowledged event", protocol.Acknowledged{})
	d.PrintJSON("Example encoding for an acknowledged event", protocol.Acknowledged{
		Type:      protocol.AcknowledgedT,
		RequestId: nextRequestId(),
	})

	d.PrintTypescriptIface("Example typescript interface stub", protocol.Acknowledged{})

	d.Printf(`

The ack event is send as a response from the receiver of a transaction in which the optional _requestId_ property has been set by the sender.
It can be used to debounce UI elements but can also be omitted to improve latency or in situations where the sender is not interested if the event has been received.
It must not enveloped into a transaction.

`)
}
