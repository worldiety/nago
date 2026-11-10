/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

/**
 * WebKit allows window.open only while a user gesture is on the call stack. A command of the backend arrives in a
 * websocket message, where this is not the case anymore, even if the user clicked just a few milliseconds ago. Other
 * browsers accept a time window after the click instead and need nothing of this.
 *
 * WebKit however forwards the gesture for up to one second to each timer, which has been created directly within that
 * gesture, but neither to nested timers nor to repeated intervals. So each trusted gesture creates many sibling timers
 * and a command, which needs the gesture, waits for the next of them. This relies on an unspecified behaviour of
 * WebKit. If it ever changes, the actions just run without a gesture, as they would without this bridge.
 */
const bridgeDelaysMs: number[] = [];
for (let delay = 25; delay <= 300; delay += 25) {
	bridgeDelaysMs.push(delay);
}
for (let delay = 350; delay <= 950; delay += 50) {
	bridgeDelaysMs.push(delay);
}

let armedUntil = 0;
let timers: ReturnType<typeof setTimeout>[] = [];
let pendingActions: (() => void)[] = [];

function runPendingActions(): void {
	const actions = pendingActions;
	pendingActions = [];
	for (const action of actions) {
		try {
			action();
		} catch (error) {
			console.error('action of a user gesture failed', error);
		}
	}
}

function arm(event: Event): void {
	if (!event.isTrusted) {
		return;
	}

	// pending actions stay queued, because the timers of this newer gesture run them just as well
	timers.forEach(clearTimeout);
	armedUntil = performance.now() + bridgeDelaysMs[bridgeDelaysMs.length - 1];
	timers = bridgeDelaysMs.map((delay) => setTimeout(runPendingActions, delay));
}

/**
 * installUserGestureBridge starts to observe user gestures, but only in browsers based on WebKit. Their vendor is
 * always Apple, including all browsers on iOS.
 */
export function installUserGestureBridge(): void {
	if (navigator.vendor !== 'Apple Computer, Inc.') {
		return;
	}

	window.addEventListener('click', arm, true);
	window.addEventListener('keydown', arm, true);
}

/**
 * withUserGesture runs an action, which the browser only allows within a user gesture and which has been requested by
 * the backend. It runs synchronously, unless the bridge is armed. Then it runs within the next timer of the gesture,
 * which is at most 50 ms later.
 */
export function withUserGesture(action: () => void): void {
	if (performance.now() >= armedUntil) {
		action();
		return;
	}

	pendingActions.push(action);
}
