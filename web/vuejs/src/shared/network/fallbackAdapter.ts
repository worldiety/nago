/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import HttpPollAdapter from '@/shared/network/httpPollAdapter';
import type ServiceAdapter from '@/shared/network/serviceAdapter';
import WebSocketAdapter from '@/shared/network/webSocketAdapter';
import { localStorageOrNull, preferHttp, rememberHttp } from '@/shared/network/wireTransport';
import { NagoEvent } from '@/shared/proto/nprotoc_gen';

// openTimeoutMillis is the time, which the websocket has to connect, before the adapter falls back to http.
const openTimeoutMillis = 3000;

// failuresBeforeFallback is the number of failed attempts of the websocket, after which the adapter falls back to
// http at once. A proxy, which blocks websockets, usually rejects the handshake right away.
const failuresBeforeFallback = 2;

/**
 * FallbackAdapter connects through the websocket and falls back to http, see HttpPollAdapter, if the websocket
 * cannot connect, e.g. because a proxy or firewall blocks it. It remembers the fallback for a while, see
 * rememberHttp, thus the next pages start with http right away. Once connected, it keeps the transport, also when the
 * connection is lost later.
 */
export default class FallbackAdapter implements ServiceAdapter {
	private active: ServiceAdapter | null = null;
	private ready: Promise<void> | null = null;

	initialize(): Promise<void> {
		if (this.ready === null) {
			this.ready = this.start();
		}

		return this.ready;
	}

	sendEvent(evt: NagoEvent): void {
		this.active?.sendEvent(evt);
	}

	async teardown(): Promise<void> {
		await this.active?.teardown();
	}

	private start(): Promise<void> {
		if (preferHttp(window.location.search, localStorageOrNull(), Date.now())) {
			console.warn('nago: connecting over http instead of the websocket');
			return this.useHttp();
		}

		return new Promise<void>((resolve) => {
			let failures = 0;
			let settled = false;
			const fallback = () => {
				if (settled) {
					return;
				}

				settled = true;
				window.clearTimeout(timer);
				console.warn('nago: the websocket cannot connect, falling back to http');
				void webSocket.teardown();
				rememberHttp(localStorageOrNull(), Date.now());
				void this.useHttp().then(resolve);
			};

			const webSocket = new WebSocketAdapter({
				onFailedBeforeOpen: () => {
					if (++failures >= failuresBeforeFallback) {
						fallback();
					}
				},
			});

			const timer = window.setTimeout(fallback, openTimeoutMillis);
			this.active = webSocket;
			void webSocket.initialize().then(() => {
				if (settled) {
					return;
				}

				settled = true;
				window.clearTimeout(timer);
				resolve();
			});
		});
	}

	private useHttp(): Promise<void> {
		const http = new HttpPollAdapter();
		this.active = http;
		return http.initialize();
	}
}
