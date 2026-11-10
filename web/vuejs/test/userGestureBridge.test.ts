/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

// Run with: npm test (node strips the types itself, no further dependencies)
import assert from 'node:assert/strict';
import { afterEach, beforeEach, describe, mock, test } from 'node:test';

const webKit = 'Apple Computer, Inc.';
let nextModule = 0;
let clock = 0;
let listeners: Map<string, (event: { isTrusted: boolean }) => void>;

async function load(vendor: string) {
	Object.defineProperty(globalThis, 'navigator', { configurable: true, value: { vendor } });
	const bridge = await import(`../src/shared/userGestureBridge.ts?instance=${nextModule++}`);
	bridge.installUserGestureBridge();
	return bridge;
}

function advance(ms: number) {
	clock += ms;
	mock.timers.tick(ms);
}

function gesture(type = 'click', isTrusted = true) {
	listeners.get(type)!({ isTrusted });
}

describe('user gesture bridge', () => {
	beforeEach(() => {
		clock = 0;
		listeners = new Map();
		mock.timers.enable({ apis: ['setTimeout'] });
		mock.method(performance, 'now', () => clock);
		Object.defineProperty(globalThis, 'window', {
			configurable: true,
			value: {
				addEventListener: (
					type: string,
					listener: (event: { isTrusted: boolean }) => void,
					capture: boolean
				) => {
					assert.equal(capture, true);
					listeners.set(type, listener);
				},
			},
		});
	});

	afterEach(() => {
		mock.timers.reset();
		mock.restoreAll();
	});

	for (const vendor of ['Google Inc.', '']) {
		test(`vendor "${vendor}" observes no gesture and runs an action at once`, async () => {
			const { withUserGesture } = await load(vendor);
			assert.equal(listeners.size, 0);

			let ran = false;
			withUserGesture(() => (ran = true));
			assert.equal(ran, true);
		});
	}

	test('an action without a preceding gesture runs at once', async () => {
		const { withUserGesture } = await load(webKit);

		let ran = false;
		withUserGesture(() => (ran = true));
		assert.equal(ran, true);
	});

	for (const type of ['click', 'keydown']) {
		test(`after a ${type} an action waits for the next timer`, async () => {
			const { withUserGesture } = await load(webKit);
			gesture(type);
			advance(10);

			let ran = false;
			withUserGesture(() => (ran = true));
			advance(14);
			assert.equal(ran, false);
			advance(1);
			assert.equal(ran, true);
		});
	}

	test('an untrusted event does not arm the bridge', async () => {
		const { withUserGesture } = await load(webKit);
		gesture('click', false);

		let ran = false;
		withUserGesture(() => (ran = true));
		assert.equal(ran, true);
	});

	test('actions of the same gesture run in their order', async () => {
		const { withUserGesture } = await load(webKit);
		gesture();

		const ran: number[] = [];
		withUserGesture(() => ran.push(1));
		withUserGesture(() => ran.push(2));
		advance(25);
		assert.deepEqual(ran, [1, 2]);
	});

	test('an action, which arrives after another one has run, gets a later timer of the same gesture', async () => {
		const { withUserGesture } = await load(webKit);
		gesture();
		advance(25);

		let ran = false;
		withUserGesture(() => (ran = true));
		assert.equal(ran, false);
		advance(25);
		assert.equal(ran, true);
	});

	test('a newer gesture keeps a pending action and runs it in its own timer', async () => {
		const { withUserGesture } = await load(webKit);
		gesture();
		advance(10);

		let ran = false;
		withUserGesture(() => (ran = true));
		advance(5);
		gesture();
		// the timer of the first gesture at 25 ms is gone
		advance(10);
		assert.equal(ran, false);
		advance(15);
		assert.equal(ran, true);
	});

	test('an action just before the end of the gesture is run by the last timer', async () => {
		const { withUserGesture } = await load(webKit);
		gesture();
		advance(949);

		let ran = false;
		withUserGesture(() => (ran = true));
		assert.equal(ran, false);
		advance(1);
		assert.equal(ran, true);
	});

	test('an action after the end of the gesture runs at once', async () => {
		const { withUserGesture } = await load(webKit);
		gesture();
		advance(950);

		let ran = false;
		withUserGesture(() => (ran = true));
		assert.equal(ran, true);
	});

	test('a failing action does not stop the following one', async () => {
		const { withUserGesture } = await load(webKit);
		mock.method(console, 'error', () => {});
		gesture();

		let ran = false;
		withUserGesture(() => {
			throw new Error('invalid url');
		});
		withUserGesture(() => (ran = true));
		advance(25);
		assert.equal(ran, true);
	});
});
