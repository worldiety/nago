/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

/*
 * webPush implements the calls of the push module. Only subscribePush registers the service worker and asks for the
 * permission, and only the backend calls it, as the reaction to an action, which the application offers for that
 * purpose. All other functions just read. Without the push module, the backend never calls any of them, and
 * installPushNavigation does nothing, because no service worker controls the page.
 *
 * The module has no imports, so that the tests can run it with node and fakes of the browser.
 */

/** PushState mirrors RetPushState of the protocol. */
export interface PushState {
	supported: boolean;
	permission: string;
	endpoint: string;
	standalone: boolean;
	requiresInstall: boolean;
}

/** PushSubscriptionData mirrors RetPushSubscription of the protocol. */
export interface PushSubscriptionData {
	endpoint: string;
	p256dh: string;
	auth: string;
	origin: string;
	userAgent: string;
}

/** SubscribeResult is the subscription, or the state, if the user has not allowed it, or an error. */
export type SubscribeResult =
	{ subscription: PushSubscriptionData } | { state: PushState } | { error: string; timeout?: boolean };

/**
 * SUBSCRIBE_TIMEOUT_MS limits the registration of the service worker and the subscription at the push service of the
 * browser, after the user has allowed the notifications. Without it, the user would wait forever, if the push service
 * of the device does not answer. It does not limit the decision of the user.
 */
export const SUBSCRIBE_TIMEOUT_MS = 30_000;

/**
 * STATE_TIMEOUT_MS limits the lookup of the subscription for the state, which must not keep a view from showing
 * the state, if the push service of the device does not answer.
 */
export const STATE_TIMEOUT_MS = 5_000;

/** PushTimeoutError tells, that the push service of the browser has not answered in time. */
export class PushTimeoutError extends Error {}

function withTimeout<T>(p: Promise<T>, ms: number): Promise<T> {
	return new Promise((resolve, reject) => {
		const timer = setTimeout(
			() => reject(new PushTimeoutError(`the push service has not answered within ${ms} ms`)),
			ms
		);
		p.then(
			(v) => {
				clearTimeout(timer);
				resolve(v);
			},
			(e) => {
				clearTimeout(timer);
				reject(e);
			}
		);
	});
}

/** GestureOps decouples the module from the user gesture bridge and the dialog of the frontend. */
export interface GestureOps {
	/** hasGesture reports, whether the browser would still accept a permission request without a new gesture. */
	hasGesture(): boolean;
	/** withGesture runs the action within the user gesture, see userGestureBridge. */
	withGesture(action: () => void): void;
	/**
	 * confirm asks the user by a dialog of the frontend and invokes allow synchronously within the click on its
	 * button, which is a new user gesture, or cancel.
	 */
	confirm(allow: () => void, cancel: () => void): void;
}

/** OPEN_MESSAGE is posted by the service worker, see serviceworker/push.js. */
export const OPEN_MESSAGE = 'nago.push.open';

/** SCOPE is the scope of the service worker, which covers the whole application. */
const SCOPE = '/';

/** pushSupported reports, whether the browser has all the interfaces, which web push requires. */
export function pushSupported(): boolean {
	return (
		typeof navigator !== 'undefined' &&
		'serviceWorker' in navigator &&
		typeof window !== 'undefined' &&
		'PushManager' in window &&
		'Notification' in window
	);
}

function isStandalone(): boolean {
	const standaloneMedia =
		typeof window.matchMedia === 'function' && window.matchMedia('(display-mode: standalone)').matches;
	return standaloneMedia || (navigator as unknown as { standalone?: boolean }).standalone === true;
}

/** isAppleMobile reports iOS and iPadOS, whose browsers receive web push only as an installed web app. */
function isAppleMobile(): boolean {
	const ua = navigator.userAgent || '';
	return /iPad|iPhone|iPod/.test(ua) || (/Macintosh/.test(ua) && (navigator.maxTouchPoints ?? 0) > 1);
}

/** currentSubscription returns the subscription of the service worker of nago, without registering anything. */
async function currentSubscription(): Promise<PushSubscription | null> {
	const registration = await navigator.serviceWorker.getRegistration(SCOPE);
	if (!registration || !registration.pushManager) {
		return null;
	}

	return registration.pushManager.getSubscription();
}

/** readPushState reads the state of the browser. It never asks the user and never registers a service worker. */
export async function readPushState(): Promise<PushState> {
	const supported = pushSupported();
	const standalone = isStandalone();
	const state: PushState = {
		supported,
		permission: supported ? Notification.permission : '',
		endpoint: '',
		standalone,
		requiresInstall: !supported && !standalone && isAppleMobile(),
	};

	if (supported) {
		try {
			// without an answer, the subscription is unknown, which a subscription settles
			state.endpoint = (await withTimeout(currentSubscription(), STATE_TIMEOUT_MS))?.endpoint ?? '';
		} catch (e) {
			console.warn('cannot read the push subscription', e);
			// e.g. a private window of Firefox has no service workers, while a push service, which does not answer,
			// still lets the user try to subscribe
			if (!(e instanceof PushTimeoutError)) {
				state.supported = false;
				state.permission = '';
			}
		}
	}

	return state;
}

/** base64UrlToBytes decodes the application server key. */
export function base64UrlToBytes(s: string): Uint8Array {
	const base64 = s.replace(/-/g, '+').replace(/_/g, '/') + '='.repeat((4 - (s.length % 4)) % 4);
	const raw = atob(base64);
	const out = new Uint8Array(raw.length);
	for (let i = 0; i < raw.length; i++) {
		out[i] = raw.charCodeAt(i);
	}

	return out;
}

function sameKey(a: ArrayBuffer | null | undefined, b: Uint8Array): boolean {
	if (!a) {
		return false;
	}

	const bytes = new Uint8Array(a);
	return bytes.length === b.length && bytes.every((v, i) => v === b[i]);
}

function toData(subscription: PushSubscription): PushSubscriptionData {
	const json = subscription.toJSON();
	return {
		endpoint: subscription.endpoint,
		p256dh: json.keys?.p256dh ?? '',
		auth: json.keys?.auth ?? '',
		origin: location.origin,
		userAgent: navigator.userAgent,
	};
}

/** activeRegistration registers the service worker and waits, until it is active, which subscribe requires. */
async function activeRegistration(serviceWorker: string): Promise<ServiceWorkerRegistration> {
	const registration = await navigator.serviceWorker.register(serviceWorker, { scope: SCOPE });
	if (registration.active) {
		return registration;
	}

	const worker = registration.installing ?? registration.waiting;
	if (worker) {
		await new Promise<void>((resolve) => {
			const check = () => {
				if (worker.state === 'activated' || worker.state === 'redundant') {
					worker.removeEventListener('statechange', check);
					resolve();
				}
			};
			worker.addEventListener('statechange', check);
			check();
		});
	}

	return registration;
}

async function subscribeGranted(applicationServerKey: string, serviceWorker: string): Promise<SubscribeResult> {
	const key = base64UrlToBytes(applicationServerKey);
	const registration = await activeRegistration(serviceWorker);

	let subscription = await registration.pushManager.getSubscription();
	if (subscription && !sameKey(subscription.options?.applicationServerKey, key)) {
		// the key of the server has changed, so the old subscription cannot be used anymore
		await subscription.unsubscribe();
		subscription = null;
	}

	if (!subscription) {
		subscription = await registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey: key,
		});
	}

	return { subscription: toData(subscription) };
}

function requestPermission(): Promise<NotificationPermission> {
	if (Notification.permission === 'granted') {
		return Promise.resolve('granted');
	}

	return Promise.resolve(Notification.requestPermission());
}

/**
 * subscribePush asks for the permission, if the user has not decided yet, registers the service worker and
 * subscribes. The permission request happens synchronously within the user gesture, which led to the call of the
 * backend. If the gesture has expired meanwhile, e.g. by a slow long polling, the dialog of the frontend asks once
 * more.
 */
export function subscribePush(
	applicationServerKey: string,
	serviceWorker: string,
	gesture: GestureOps,
	timeoutMs = SUBSCRIBE_TIMEOUT_MS
): Promise<SubscribeResult> {
	return new Promise((resolve) => {
		const finish = (p: Promise<SubscribeResult>) => {
			p.then(resolve, (e) =>
				resolve(e instanceof PushTimeoutError ? { error: e.message, timeout: true } : { error: String(e) })
			);
		};

		const unsubscribed = () => readPushState().then((state) => ({ state }) as SubscribeResult);

		if (!pushSupported() || Notification.permission === 'denied') {
			finish(unsubscribed());
			return;
		}

		// must be called synchronously within the gesture, i.e. before any await
		const run = () => {
			finish(
				requestPermission().then((permission) =>
					permission === 'granted'
						? withTimeout(subscribeGranted(applicationServerKey, serviceWorker), timeoutMs)
						: unsubscribed()
				)
			);
		};

		if (Notification.permission === 'granted') {
			run();
			return;
		}

		if (gesture.hasGesture()) {
			gesture.withGesture(run);
			return;
		}

		gesture.confirm(run, () => finish(unsubscribed()));
	});
}

/**
 * unsubscribePush removes the subscription and returns it, or undefined, if there was none. The service worker
 * stays registered and the permission stays as it is.
 */
export async function unsubscribePush(): Promise<PushSubscriptionData | undefined> {
	if (!pushSupported()) {
		return undefined;
	}

	const subscription = await withTimeout(currentSubscription(), SUBSCRIBE_TIMEOUT_MS);
	if (!subscription) {
		return undefined;
	}

	const data = toData(subscription);
	await subscription.unsubscribe();
	return data;
}

let navigationInstalled = false;

/**
 * installPushNavigation lets the service worker open a route in this window by a click on a notification. At the
 * start, it only listens, if a service worker controls the page, which only happens after a subscription by the
 * push module. After a subscription, the caller passes subscribed, because the new service worker may take control
 * only a little later.
 */
export function installPushNavigation(navigate: (url: string) => void, subscribed = false): void {
	if (navigationInstalled || typeof navigator === 'undefined' || !('serviceWorker' in navigator)) {
		return;
	}

	if (!subscribed && !navigator.serviceWorker.controller) {
		return;
	}

	navigationInstalled = true;
	navigator.serviceWorker.addEventListener('message', (event: MessageEvent) => {
		const data = event.data;
		if (!data || data.type !== OPEN_MESSAGE || typeof data.url !== 'string') {
			return;
		}

		// the service worker sends only paths of its own origin, but a message must not lead anywhere else
		const url = new URL(data.url, location.origin);
		if (url.origin !== location.origin) {
			return;
		}

		navigate(url.pathname + url.search + url.hash);
	});

	// messages are queued until the document has loaded, unless started explicitly
	navigator.serviceWorker.startMessages();
}
