/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

/**
 * TextFieldSync decides, whether a text field adopts the value of a server rendering or keeps the text the user has
 * typed. All values are parsed values, i.e. in the representation which is transmitted to the server.
 *
 * Its anchor is the value which the server has or is about to have, as far as the text field knows: the value it has
 * transmitted last or the server value it has adopted last.
 *
 * As long as the input equals the anchor, the server knows everything the user has typed and a rendering carries the
 * newer value. This is what makes a server side reset work while the field keeps the focus, e.g. a chat input which
 * is emptied by its own submit handler.
 *
 * If the input differs from the anchor, the user has typed something which has not been transmitted yet, typically
 * because the debounce is still running. No rendering can carry that text, whatever caused it: the server renders on
 * its own, e.g. because a goroutine changed a state (such a rendering has the request tracing id 0, which App.vue
 * never discards as outdated), or another component caused it. The input wins then, and the pending debounce
 * transmits it.
 *
 * A transmitted value is in flight until a rendering shows, that the server has processed it: one which answers its
 * request or a later one (see renderedRID in eventhandling.ts), or one which carries it. Meanwhile, a rendering with
 * another value may have been made before the server received ours, typically a rendering the server made on its own
 * while our value was on its way. Adopting it would throw the field back to an older value, and whatever the user
 * typed in the meantime would continue from there. Thus, such renderings are ignored as well.
 *
 * This module has no dependencies, so that node can test it directly, see test/textFieldSync.test.ts.
 */
export class TextFieldSync {
	private anchor: string;
	// inFlight is the request tracing id, which has transmitted the anchor, as long as no rendering has shown that
	// the server has processed it, otherwise 0.
	private inFlight: number = 0;

	constructor(initialValue: string) {
		this.anchor = initialValue;
	}

	/**
	 * hasUnsentInput tells, whether the user has typed something, which the server does not know, so that it must be
	 * transmitted.
	 */
	hasUnsentInput(input: string): boolean {
		return input !== this.anchor;
	}

	/**
	 * sent records, that the value has been transmitted to the server with the given request tracing id.
	 */
	sent(value: string, rid: number): void {
		this.anchor = value;
		this.inFlight = rid;
	}

	/**
	 * adopt tells, whether the text field replaces its input with the server value of a rendering. If so, the server
	 * value becomes the anchor. renderedRID is the highest request tracing id answered by a rendering so far.
	 */
	adopt(input: string, serverValue: string, renderedRID: number): boolean {
		if (this.hasUnsentInput(input)) {
			return false;
		}

		if (serverValue === this.anchor) {
			this.inFlight = 0; // the server has it
			return false;
		}

		if (this.inFlight > renderedRID) {
			return false; // possibly made before the server received our value
		}

		this.inFlight = 0;
		this.anchor = serverValue;
		return true;
	}
}
