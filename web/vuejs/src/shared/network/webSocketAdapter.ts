/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { v4 as uuidv4 } from 'uuid';
import ConnectionHandler from '@/shared/network/connectionHandler';
import type ServiceAdapter from '@/shared/network/serviceAdapter';
import { BinaryReader, BinaryWriter, NagoEvent, Ping, marshal, unmarshal } from '@/shared/proto/nprotoc_gen';

export default class WebSocketAdapter implements ServiceAdapter {
	private readonly isSecure: boolean = false;
	private readonly scopeId: string;
	private webSocket: WebSocket | null = null;
	private closedGracefully: boolean = false;
	private retryTimeout: number | null = null;
	private retries: number = 0;
	private lastEventSendAt: number;

	constructor() {
		this.lastEventSendAt = new Date().getTime();
		// important: keep this scopeId for the resume capability only once per
		// channel. Otherwise, (e.g. when storing in localstorage or cookie) all
		// browser tabs and windows will try to steal the scope from each other.
		// So, you MUST ensure that each VueJS instance has its own unique scope id,
		// also when reconnecting to an existing scope.
		this.scopeId = uuidv4();
		this.isSecure = location.protocol == 'https:';
	}

	async initialize(): Promise<void> {
		let proto = 'ws';
		if (this.isSecure) {
			proto = 'wss';
		}
		let webSocketURL = `${proto}://${window.location.hostname}:${window.location.port}/wire?_sid=${this.scopeId}`;
		const queryString = window.location.search.substring(1);
		if (queryString) {
			webSocketURL += `&${queryString}`;
		}

		return new Promise<void>((resolve) => {
			this.webSocket = new WebSocket(webSocketURL);
			this.webSocket.binaryType = 'arraybuffer';

			this.webSocket.onmessage = (e) => this.receiveBinary(e.data);

			this.webSocket.onclose = (evt) => {
				ConnectionHandler.connectionChanged({ connected: false });

				if (this.closedGracefully) {
					// Try to reopen the socket if it was not closed gracefully
					window.console.log('ws was intentionally closed');
				} else {
					// Keep the socket closed if it was closed gracefully (i.e. intentional)
					window.console.log('ws was not closed gracefully');
					console.log('WebSocket closed:', evt);
					console.log('Code:', evt.code);
					console.log('Reason:', evt.reason);
					console.log('Was clean:', evt.wasClean);

					this.closedGracefully = false;
					this.retry();
				}
			};

			this.webSocket.onerror = (ev) => {
				window.console.log('websocket failed', ev);
			};

			this.webSocket.onopen = () => {
				ConnectionHandler.connectionChanged({ connected: true });
				this.retries = 0;

				// this keeps our connection at least logically alive
				setInterval(() => {
					if (this.closedGracefully) {
						return;
					}

					this.sendEvent(new Ping());
				}, 30000);

				resolve();
			};
		});
	}

	sendEvent(evt: NagoEvent): void {
		const startTime = new Date().getTime();
		//console.log('SEND EVENT', evt);
		let writer = new BinaryWriter();
		marshal(writer, evt);
		let buffer = writer.getBuffer();
		//console.log('nprotoc buffer', buffer);
		const endTime = new Date().getTime();
		console.log(`nprotoc marshal time ${endTime - startTime} ms. type =`, evt.constructor.name);
		this.webSocket?.send(buffer);
		this.lastEventSendAt = new Date().getTime();
	}

	async teardown(): Promise<void> {
		window.console.log('websocket teardown');
		this.closedGracefully = true;
		this.webSocket?.close();
		ConnectionHandler.connectionChanged({ connected: false });
	}

	private retry() {
		console.log('retry');
		if (this.retryTimeout !== null) {
			return;
		}

		let timeout = 50; // Retry timeout 50ms
		if (this.retries >= 20) {
			timeout = 250; // 250ms after 2s
		}
		if (this.retries >= 32) {
			timeout = 1000; // 1s after 5s
		}
		if (this.retries >= 37) {
			timeout = 2000; // 2s after 10s
		}

		this.retries += 1;
		this.retryTimeout = window.setTimeout(() => {
			this.retryTimeout = null;
			this.initialize();
		}, timeout);
	}

	private receiveBinary(responseRaw: ArrayBuffer): void {
		let endTime = new Date().getTime();
		let responseTime = endTime - this.lastEventSendAt;

		const startTime = new Date().getTime();
		//console.log('WS received', responseRaw);
		let msg = unmarshal(new BinaryReader(new Uint8Array(responseRaw))); // TODO i don't know what i'm doing here, does it copy?
		endTime = new Date().getTime();
		console.log(
			`nprotoc response after ${responseTime}ms, unmarshal time ${endTime - startTime}ms, total ${responseTime + endTime - startTime}ms, type =`,
			msg.constructor.name
		);
		ConnectionHandler.publishEvent(msg as NagoEvent);
	}
}
