/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

/**
 * transportParam selects the transport of the wire by the query of the page: _transport=http skips the websocket,
 * e.g. to diagnose a proxy. The parameter is not passed to the application.
 */
export const transportParam = '_transport';

/**
 * rememberHttpMillis is the duration, for which a page starts with http after the websocket has failed, instead of
 * waiting for the websocket again.
 */
export const rememberHttpMillis = 24 * 60 * 60 * 1000;

const storageKey = 'nago.wire.http-until';

/**
 * preferHttp tells whether the wire skips the websocket, because the query of the page requests it or the websocket
 * has failed recently.
 */
export function preferHttp(search: string, storage: Storage | null, now: number): boolean {
	if (new URLSearchParams(search).get(transportParam) === 'http') {
		return true;
	}

	try {
		const until = Number(storage?.getItem(storageKey) ?? 0);
		return until > now;
	} catch {
		return false;
	}
}

/**
 * rememberHttp lets the pages of the next rememberHttpMillis start with http.
 */
export function rememberHttp(storage: Storage | null, now: number): void {
	try {
		storage?.setItem(storageKey, String(now + rememberHttpMillis));
	} catch {
		// the storage is unavailable, e.g. in a private window, thus the next page tries the websocket again
	}
}

/**
 * withoutTransportParam removes transportParam from the query of the page, without the leading question mark.
 */
export function withoutTransportParam(query: string): string {
	const params = new URLSearchParams(query);
	if (!params.has(transportParam)) {
		return query;
	}

	params.delete(transportParam);
	return params.toString();
}

/**
 * localStorageOrNull returns the local storage, unless the browser denies it.
 */
export function localStorageOrNull(): Storage | null {
	try {
		return window.localStorage;
	} catch {
		return null;
	}
}
