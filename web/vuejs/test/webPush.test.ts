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
import { afterEach, beforeEach, describe, test } from 'node:test';

const SW = '/api/nago/v1/push/sw.js';
const KEY = 'BEl62iUYgUivxIkv69yViEuiBIa-Ib9-SkvMeAtA3LFgDzkrxZJjSgSnfckjBJuBkr3qBUYIHBQFLXYp5Nksh8U';

let nextModule = 0;
let log: string[];
let permission: string;
let promptAnswer: string;
let inGesture: boolean;
let registration: any;
let existing: any;
let messageListener: ((event: { data: any }) => void) | undefined;

function fakeSubscription(endpoint: string, key: Uint8Array) {
	return {
		endpoint,
		options: { applicationServerKey: key.buffer },
		toJSON: () => ({ endpoint, keys: { p256dh: 'p256dh-' + endpoint, auth: 'auth-' + endpoint } }),
		unsubscribe: async () => {
			log.push('unsubscribe ' + endpoint);
			existing = null;
			return true;
		},
	};
}

function install(opts: { push?: boolean; controller?: boolean; userAgent?: string; standalone?: boolean } = {}) {
	const pushSupported = opts.push ?? true;
	registration = {
		active: { state: 'activated' },
		pushManager: {
			getSubscription: async () => existing,
			subscribe: async (options: { userVisibleOnly: boolean; applicationServerKey: Uint8Array }) => {
				log.push('subscribe ' + options.userVisibleOnly);
				existing = fakeSubscription('https://push.example/new', options.applicationServerKey);
				return existing;
			},
		},
	};

	const serviceWorker = {
		controller: opts.controller ? {} : null,
		getRegistration: async (scope: string) => {
			log.push('getRegistration ' + scope);
			return existing === undefined ? undefined : registration;
		},
		register: async (url: string, options: { scope: string }) => {
			log.push(`register ${url} ${options.scope}`);
			return registration;
		},
		addEventListener: (type: string, listener: any) => {
			log.push('listen ' + type);
			messageListener = listener;
		},
		startMessages: () => log.push('startMessages'),
	};

	const notification = {
		get permission() {
			return permission;
		},
		requestPermission: () => {
			log.push('requestPermission' + (inGesture ? ' in gesture' : ' without gesture'));
			permission = promptAnswer;
			return Promise.resolve(promptAnswer);
		},
	};

	const win: any = {
		matchMedia: () => ({ matches: opts.standalone ?? false }),
	};
	if (pushSupported) {
		win.PushManager = function () {};
		win.Notification = notification;
	}

	Object.defineProperty(globalThis, 'window', { configurable: true, value: win });
	Object.defineProperty(globalThis, 'Notification', { configurable: true, value: notification });
	Object.defineProperty(globalThis, 'location', { configurable: true, value: { origin: 'https://app.example' } });
	Object.defineProperty(globalThis, 'navigator', {
		configurable: true,
		value: {
			userAgent: opts.userAgent ?? 'Mozilla/5.0 (Macintosh) Chrome/140',
			maxTouchPoints: 0,
			...(pushSupported || opts.userAgent ? { serviceWorker } : {}),
		},
	});
}

async function load() {
	return await import(`../src/shared/webPush.ts?instance=${nextModule++}`);
}

/** gesture returns fake operations, which record, how the permission request is reached. */
function gesture(has: boolean, confirmAnswer?: 'allow' | 'cancel') {
	return {
		hasGesture: () => has,
		withGesture: (action: () => void) => {
			log.push('withGesture');
			inGesture = true;
			try {
				action();
			} finally {
				inGesture = false;
			}
		},
		confirm: (allow: () => void, cancel: () => void) => {
			log.push('confirm');
			if (confirmAnswer === 'allow') {
				inGesture = true;
				try {
					allow();
				} finally {
					inGesture = false;
				}
			} else if (confirmAnswer === 'cancel') {
				cancel();
			}
		},
	};
}

/** asking reports the calls, which may let a browser ask the user or register the service worker. */
function asking(): string[] {
	return log.filter((l) => /^(requestPermission|register|subscribe|confirm)/.test(l));
}

describe('web push', () => {
	beforeEach(() => {
		log = [];
		permission = 'default';
		promptAnswer = 'granted';
		inGesture = false;
		existing = undefined;
		messageListener = undefined;
	});

	afterEach(() => {
		for (const name of ['window', 'Notification', 'location', 'navigator']) {
			delete (globalThis as any)[name];
		}
	});

	test('reading the state never asks and never registers', async () => {
		install();
		const push = await load();
		const state = await push.readPushState();
		assert.deepEqual(state, {
			supported: true,
			permission: 'default',
			endpoint: '',
			standalone: false,
			requiresInstall: false,
		});

		existing = fakeSubscription('https://push.example/old', push.base64UrlToBytes(KEY));
		assert.equal((await push.readPushState()).endpoint, 'https://push.example/old');
		assert.deepEqual(asking(), []);
	});

	test('Safari on iOS requires the installation', async () => {
		install({ push: false, userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) Safari' });
		const push = await load();
		const state = await push.readPushState();
		assert.equal(state.supported, false);
		assert.equal(state.requiresInstall, true);

		const result = await push.subscribePush(KEY, SW, gesture(true));
		assert.ok('state' in result);
		assert.deepEqual(asking(), []);
	});

	test('a browser without push is not supported', async () => {
		install({ push: false });
		const push = await load();
		assert.deepEqual(await push.readPushState(), {
			supported: false,
			permission: '',
			endpoint: '',
			standalone: false,
			requiresInstall: false,
		});
	});

	test('subscribing asks within the gesture, then registers and subscribes', async () => {
		install();
		const push = await load();
		const result = await push.subscribePush(KEY, SW, gesture(true));

		assert.deepEqual(asking(), ['requestPermission in gesture', `register ${SW} /`, 'subscribe true']);
		assert.deepEqual(result, {
			subscription: {
				endpoint: 'https://push.example/new',
				p256dh: 'p256dh-https://push.example/new',
				auth: 'auth-https://push.example/new',
				origin: 'https://app.example',
				userAgent: 'Mozilla/5.0 (Macintosh) Chrome/140',
			},
		});
		assert.deepEqual([...push.base64UrlToBytes(KEY)].slice(0, 2), [0x04, 0x49], 'the key is an uncompressed point');
	});

	test('an expired gesture asks by the dialog of the frontend first', async () => {
		install();
		const push = await load();
		const result = await push.subscribePush(KEY, SW, gesture(false, 'allow'));
		assert.deepEqual(asking(), ['confirm', 'requestPermission in gesture', `register ${SW} /`, 'subscribe true']);
		assert.ok('subscription' in result);
	});

	test('cancelling the dialog asks nothing', async () => {
		install();
		const push = await load();
		const result = await push.subscribePush(KEY, SW, gesture(false, 'cancel'));
		assert.deepEqual(asking(), ['confirm']);
		assert.ok('state' in result && result.state.permission === 'default');
	});

	test('a denied prompt neither registers nor subscribes', async () => {
		install();
		promptAnswer = 'denied';
		const push = await load();
		const result = await push.subscribePush(KEY, SW, gesture(true));
		assert.deepEqual(asking(), ['requestPermission in gesture']);
		assert.ok('state' in result && result.state.permission === 'denied');
	});

	test('a denied permission asks nothing', async () => {
		install();
		permission = 'denied';
		const push = await load();
		const result = await push.subscribePush(KEY, SW, gesture(true));
		assert.deepEqual(asking(), []);
		assert.ok('state' in result && result.state.permission === 'denied');
	});

	test('a push service without answer ends in an error instead of waiting forever', async () => {
		install();
		const push = await load();
		registration.pushManager.subscribe = () => {
			log.push('subscribe hangs');
			return new Promise(() => {});
		};

		const result = await push.subscribePush(KEY, SW, gesture(true), 20);
		assert.deepEqual(asking(), ['requestPermission in gesture', `register ${SW} /`, 'subscribe hangs']);
		assert.ok('error' in result && result.timeout === true, JSON.stringify(result));
	});

	test('a push service without answer does not block the state', async () => {
		install();
		const push = await load();
		existing = fakeSubscription('https://push.example/old', push.base64UrlToBytes(KEY));
		registration.pushManager.getSubscription = () => new Promise(() => {});

		const started = Date.now();
		const state = await Promise.race([
			push.readPushState(),
			new Promise((_, reject) =>
				setTimeout(() => reject(new Error('state blocks')), push.STATE_TIMEOUT_MS + 1000)
			),
		]);
		assert.equal(state.supported, true);
		assert.equal(state.endpoint, '');
		assert.ok(Date.now() - started >= push.STATE_TIMEOUT_MS - 50);
	});

	test('the time of the user to decide is not limited', async () => {
		install();
		const push = await load();
		let decide: (p: string) => void = () => {};
		(globalThis as any).Notification.requestPermission = () => {
			log.push('requestPermission');
			return new Promise((resolve) => (decide = resolve));
		};

		const pending = push.subscribePush(KEY, SW, gesture(true), 20);
		await new Promise((r) => setTimeout(r, 60));
		permission = 'granted';
		decide('granted');
		const result = await pending;
		assert.ok('subscription' in result, JSON.stringify(result));
	});

	test('a granted permission needs no gesture', async () => {
		install();
		permission = 'granted';
		const push = await load();
		const result = await push.subscribePush(KEY, SW, gesture(false));
		assert.deepEqual(asking(), [`register ${SW} /`, 'subscribe true']);
		assert.ok('subscription' in result);
	});

	test('a subscription for another key is replaced', async () => {
		install();
		permission = 'granted';
		const push = await load();
		existing = fakeSubscription('https://push.example/old', new Uint8Array([1, 2, 3]));
		await push.subscribePush(KEY, SW, gesture(true));
		assert.ok(log.includes('unsubscribe https://push.example/old'));
		assert.ok(log.includes('subscribe true'));

		// the same key keeps the subscription
		log = [];
		await push.subscribePush(KEY, SW, gesture(true));
		assert.deepEqual(asking(), [`register ${SW} /`]);
	});

	test('unsubscribing returns the removed subscription and registers nothing', async () => {
		install();
		const push = await load();
		assert.equal(await push.unsubscribePush(), undefined);

		existing = fakeSubscription('https://push.example/old', push.base64UrlToBytes(KEY));
		const removed = await push.unsubscribePush();
		assert.equal(removed.endpoint, 'https://push.example/old');
		assert.deepEqual(asking(), []);
		assert.ok(log.includes('unsubscribe https://push.example/old'));
	});

	test('the navigation listens only with a service worker of nago', async () => {
		install();
		let push = await load();
		const navigated: string[] = [];
		push.installPushNavigation((url: string) => navigated.push(url));
		assert.equal(messageListener, undefined, 'without a controlling service worker nothing listens');

		install({ controller: true });
		push = await load();
		push.installPushNavigation((url: string) => navigated.push(url));
		assert.ok(messageListener);
		messageListener!({ data: { type: 'nago.push.open', url: '/orders?id=42#top' } });
		messageListener!({ data: { type: 'nago.push.open', url: 'https://evil.example/x' } });
		messageListener!({ data: { type: 'other', url: '/x' } });
		assert.deepEqual(navigated, ['/orders?id=42#top']);

		// right after a subscription, the new service worker may not control the page yet
		install();
		push = await load();
		push.installPushNavigation(() => {}, true);
		assert.ok(log.includes('listen message') && log.includes('startMessages'));
	});
});
