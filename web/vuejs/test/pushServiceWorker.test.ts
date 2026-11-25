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
import { readFileSync } from 'node:fs';
import { beforeEach, describe, test } from 'node:test';

const source = readFileSync(new URL('../serviceworker/push.js', import.meta.url), 'utf8');

interface FakeWindow {
	url: string;
	focused: boolean;
	visibilityState: string;
	messages: any[];
	focusCalls: number;
	postMessage(msg: any): void;
	focus(): Promise<FakeWindow>;
}

let handlers: Map<string, (event: any) => void>;
let shown: { title: string; options: any }[];
let opened: string[];
let windows: FakeWindow[];
let log: string[];

function fakeWindow(url: string, focused = false): FakeWindow {
	return {
		url,
		focused,
		visibilityState: focused ? 'visible' : 'hidden',
		messages: [],
		focusCalls: 0,
		postMessage(msg: any) {
			this.messages.push(msg);
		},
		async focus() {
			this.focusCalls++;
			return this;
		},
	};
}

function run() {
	const self = {
		location: { origin: 'https://app.example', host: 'app.example' },
		addEventListener: (type: string, fn: (event: any) => void) => handlers.set(type, fn),
		skipWaiting: () => log.push('skipWaiting'),
		registration: {
			showNotification: async (title: string, options: any) => {
				shown.push({ title, options });
			},
		},
		clients: {
			claim: async () => log.push('claim'),
			matchAll: async (options: any) => {
				assert.deepEqual(options, { type: 'window', includeUncontrolled: true });
				return windows;
			},
			openWindow: async (url: string) => {
				opened.push(url);
				return null;
			},
		},
	};

	new Function('self', source)(self);
}

/** dispatch runs a handler and waits for its waitUntil. */
async function dispatch(type: string, event: any) {
	let pending: Promise<unknown> = Promise.resolve();
	handlers.get(type)!({ ...event, waitUntil: (p: Promise<unknown>) => (pending = p) });
	await pending;
}

function pushEvent(data: unknown) {
	if (data === undefined) {
		return { data: null };
	}

	const text = typeof data === 'string' ? data : JSON.stringify(data);
	return {
		data: {
			json: () => JSON.parse(text),
			text: () => text,
		},
	};
}

function clickEvent(url: string) {
	let closed = false;
	return {
		notification: {
			data: { url },
			close: () => (closed = true),
			get closed() {
				return closed;
			},
		},
	};
}

describe('push service worker', () => {
	beforeEach(() => {
		handlers = new Map();
		shown = [];
		opened = [];
		windows = [];
		log = [];
		run();
	});

	test('it has no fetch handler, so that requests go past it', () => {
		assert.deepEqual([...handlers.keys()].sort(), ['activate', 'install', 'notificationclick', 'push']);
	});

	test('it activates at once', async () => {
		handlers.get('install')!({});
		await dispatch('activate', {});
		assert.deepEqual(log, ['skipWaiting', 'claim']);
	});

	test('a push shows the notification of the message', async () => {
		await dispatch(
			'push',
			pushEvent({
				title: 'Neuer Auftrag',
				body: 'Auftrag 42',
				icon: '/icon.png',
				tag: 'order-42',
				url: '/orders?id=42',
			})
		);
		assert.deepEqual(shown, [
			{
				title: 'Neuer Auftrag',
				options: {
					body: 'Auftrag 42',
					icon: '/icon.png',
					tag: 'order-42',
					renotify: true,
					data: { url: '/orders?id=42' },
				},
			},
		]);
	});

	test('a push never links to another site and always shows something', async () => {
		await dispatch('push', pushEvent({ title: 'x', url: 'https://evil.example/phish' }));
		await dispatch('push', pushEvent('plain text'));
		await dispatch('push', pushEvent(undefined));
		assert.equal(shown[0].options.data.url, '/');
		assert.equal(shown[1].title, 'app.example');
		assert.equal(shown[1].options.body, 'plain text');
		assert.equal(shown[2].title, 'app.example');
	});

	test('a click navigates the focused window of the application', async () => {
		const other = fakeWindow('https://other.example/', true);
		const background = fakeWindow('https://app.example/a');
		const focused = fakeWindow('https://app.example/b', true);
		windows = [other, background, focused];

		const event = clickEvent('/orders?id=42');
		await dispatch('notificationclick', event);
		assert.equal(event.notification.closed, true);
		assert.deepEqual(focused.messages, [{ type: 'nago.push.open', url: '/orders?id=42' }]);
		assert.equal(focused.focusCalls, 1);
		assert.deepEqual(background.messages, []);
		assert.deepEqual(other.messages, []);
		assert.deepEqual(opened, []);
	});

	test('a click opens a window, if none is open', async () => {
		windows = [fakeWindow('https://other.example/', true)];
		await dispatch('notificationclick', clickEvent('/orders?id=42'));
		assert.deepEqual(opened, ['/orders?id=42']);
	});
});
