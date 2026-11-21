/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

/**
 * OptimisticToggle keeps a boolean, which the user switches in the browser at once, while the server takes the new
 * value with a round trip and a rendering, e.g. the open state of an accordion. The browser shows the local value
 * immediately, and the server value of a later rendering decides again.
 *
 * A toggle is in flight until a rendering shows, that the server has processed it: one which answers its request or a
 * later one (see renderedRID in eventhandling.ts). Meanwhile, renderings keep the local value, because they may have
 * been made before the server received the toggle, e.g. a rendering the server made on its own (request tracing id 0).
 *
 * Afterwards, every rendering decides, also one with an unchanged value: if the server did not take the toggle (e.g.
 * because the state has been removed from the tree in the meantime), the rendering answers the request with the old
 * value and the local value falls back to it. This is why the owner must call adopt for every new rendering and not
 * only, if the server value changes.
 *
 * This module has no dependencies, so that node can test it directly, see test/optimisticToggle.test.ts.
 */
export class OptimisticToggle {
	private local: boolean;
	// inFlight is the request tracing id of the last toggle, as long as no rendering has shown that the server has
	// processed it, otherwise 0.
	private inFlight: number = 0;
	private state: number | undefined;

	constructor(serverValue: boolean, state?: number) {
		this.local = serverValue;
		this.state = state;
	}

	/** value returns the value, which the browser shows. */
	get value(): boolean {
		return this.local;
	}

	/**
	 * toggle switches the local value at once and records the request tracing id, which transmits it.
	 */
	toggle(value: boolean, rid: number): void {
		this.local = value;
		this.inFlight = rid;
	}

	/**
	 * adopt takes the server value of a rendering, unless a toggle of the same bound state is still in flight, and
	 * returns the value to show. renderedRID is the highest request tracing id answered by a rendering so far.
	 *
	 * A stack patches its children by their position, so an instance is reused for whichever component moves to its
	 * place. If it shows another bound state than before, it starts over with the server value.
	 */
	adopt(serverValue: boolean, state: number | undefined, renderedRID: number): boolean {
		if (state !== this.state) {
			this.state = state;
			this.inFlight = 0;
			this.local = serverValue;
			return this.local;
		}

		if (this.inFlight > renderedRID) {
			return this.local; // possibly made before the server received our toggle
		}

		this.inFlight = 0;
		this.local = serverValue;
		return this.local;
	}
}
