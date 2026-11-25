/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

/*
 * The service worker of nago shows the web push notifications and opens the application by a click on one. Only
 * the push module serves it, see package cfgpush, and the frontend registers it only after the user has asked for
 * notifications. It has no fetch handler on purpose: all requests, the websocket and the long polling of /wire go
 * past it to the network as without a service worker.
 *
 * It is plain JavaScript without a build step and uses only self, so that a test can run it with a fake.
 */

/** OPEN_MESSAGE is posted to an open window of the application, which then navigates to the url by itself. */
const OPEN_MESSAGE = 'nago.push.open';

self.addEventListener('install', () => {
	// a new version replaces the old one at once, because it keeps no state
	self.skipWaiting();
});

self.addEventListener('activate', (event) => {
	event.waitUntil(self.clients.claim());
});

/**
 * parsePayload reads the message of the server. A message, which is no JSON, is shown as text, because the browser
 * requires a notification for each push.
 */
function parsePayload(data) {
	if (!data) {
		return {};
	}

	try {
		const payload = data.json();
		return payload && typeof payload === 'object' ? payload : { body: String(payload) };
	} catch (e) {
		return { body: data.text() };
	}
}

/**
 * sameOriginPath returns the path, query and fragment of the url, if it belongs to the application, and '/'
 * otherwise, so that a notification never opens another site.
 */
function sameOriginPath(url) {
	try {
		const u = new URL(url || '/', self.location.origin);
		if (u.origin !== self.location.origin) {
			return '/';
		}

		return u.pathname + u.search + u.hash;
	} catch (e) {
		return '/';
	}
}

self.addEventListener('push', (event) => {
	const payload = parsePayload(event.data);
	const options = {
		body: payload.body || '',
		data: { url: sameOriginPath(payload.url) },
	};

	if (payload.icon) {
		options.icon = payload.icon;
	}

	if (payload.tag) {
		options.tag = payload.tag;
		options.renotify = true;
	}

	event.waitUntil(self.registration.showNotification(payload.title || self.location.host, options));
});

self.addEventListener('notificationclick', (event) => {
	event.notification.close();
	const path = sameOriginPath(event.notification.data && event.notification.data.url);

	event.waitUntil(
		self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windows) => {
			const own = windows.filter((w) => {
				try {
					return new URL(w.url).origin === self.location.origin;
				} catch (e) {
					return false;
				}
			});

			// prefer the window, which the user has looked at last
			const target = own.find((w) => w.focused) || own.find((w) => w.visibilityState === 'visible') || own[0];
			if (!target) {
				return self.clients.openWindow(path);
			}

			target.postMessage({ type: OPEN_MESSAGE, url: path });
			return target.focus();
		})
	);
});
