/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { v4 as uuidv4 } from 'uuid';
import ConnectionHandler from '@/shared/network/connectionHandler';
import ServerInstance from '@/shared/network/serverInstance';
import type ServiceAdapter from '@/shared/network/serviceAdapter';
import { decodeFrames, encodeFrames } from '@/shared/network/wireFrames';
import { BinaryReader, BinaryWriter, NagoEvent, marshal, unmarshal } from '@/shared/proto/nprotoc_gen';

// channelHeader carries the token of the channel, which the backend has connected to the scope.
const channelHeader = 'X-Nago-Channel';

// sendAttempts is the number of attempts to send events, before the adapter connects anew.
const sendAttempts = 5;

/**
 * HttpPollAdapter connects to the backend over plain http, when a proxy or firewall blocks the websocket. It polls
 * the messages of the scope, which the backend answers as soon as there are some, and posts the events. Both
 * directions are numbered, thus a request, which is repeated after a network error, neither loses nor duplicates a
 * message. When the backend has connected the scope to a new channel, e.g. after a restart, the adapter reports a
 * reconnect, which makes the app render the window again.
 */
export default class HttpPollAdapter implements ServiceAdapter {
	private readonly scopeId: string;
	private readonly instance = new ServerInstance();
	private token = '';
	private ack = 0;
	private sendSeq = 0;
	private pending: Uint8Array[] = [];
	private sending = false;
	private stopped = false;
	private connected = false;
	private reconnecting = false;
	private retries = 0;
	private abort: AbortController | null = null;
	private ready: Promise<void> | null = null;
	private resolveReady: (() => void) | null = null;

	constructor() {
		// like the websocket, each instance of the app has its own scope, see WebSocketAdapter
		this.scopeId = uuidv4();
	}

	initialize(): Promise<void> {
		if (this.ready === null) {
			this.ready = new Promise<void>((resolve) => (this.resolveReady = resolve));
			void this.pollLoop();
		}

		return this.ready;
	}

	async teardown(): Promise<void> {
		this.stopped = true;
		this.abort?.abort();
		this.setConnected(false);
	}

	sendEvent(evt: NagoEvent): void {
		const writer = new BinaryWriter();
		marshal(writer, evt);
		this.pending.push(writer.getBuffer());
		void this.flush();
	}

	private async pollLoop(): Promise<void> {
		while (!this.stopped) {
			this.reconnecting = false;
			this.abort = new AbortController();
			try {
				const params = new URLSearchParams({ _sid: this.scopeId, ack: String(this.ack) });
				if (this.token !== '') {
					params.set('_c', this.token);
				}

				const response = await fetch(`/wire/poll?${params}`, {
					method: 'POST',
					cache: 'no-store',
					signal: this.abort.signal,
				});

				if (response.status === 409) {
					// the channel is gone, the next poll gets a new one
					this.token = '';
					this.setConnected(false);
					continue;
				}

				if (!response.ok) {
					throw new Error(`poll failed with status ${response.status}`);
				}

				const token = response.headers.get(channelHeader) ?? '';
				const body = new Uint8Array(await response.arrayBuffer());
				if (token !== this.token) {
					// a new channel: the events of the former one are lost, and the app requests a render again
					this.setConnected(false);
					this.token = token;
					this.ack = 0;
					this.sendSeq = 0;
					this.pending = [];
				}

				this.retries = 0;
				this.setConnected(true);
				this.receive(body);
				void this.flush();
			} catch (err) {
				if (this.stopped) {
					return;
				}

				if (this.reconnecting) {
					continue;
				}

				console.log('http wire poll failed', err);
				this.setConnected(false);
				await sleep(this.backoff());
			}
		}
	}

	private receive(body: Uint8Array): void {
		const { first, msgs } = decodeFrames(body);
		msgs.forEach((msg, i) => {
			const seq = first + i;
			if (seq <= this.ack) {
				return;
			}

			ConnectionHandler.publishEvent(unmarshal(new BinaryReader(msg)) as NagoEvent);
			this.ack = seq;
		});
	}

	// flush posts the pending events, one request at a time, which keeps them in order.
	private async flush(): Promise<void> {
		if (this.sending || this.pending.length === 0 || !this.connected) {
			return;
		}

		this.sending = true;
		const token = this.token;
		const batch = this.pending.splice(0);
		const first = this.sendSeq + 1;
		const body = encodeFrames(first, batch);
		try {
			for (let attempt = 1; ; attempt++) {
				if (this.stopped || token !== this.token) {
					return;
				}

				try {
					const params = new URLSearchParams({ _sid: this.scopeId, _c: token });
					const response = await fetch(`/wire/send?${params}`, {
						method: 'POST',
						cache: 'no-store',
						headers: { 'Content-Type': 'application/octet-stream' },
						body,
					});

					if (response.ok) {
						if (token === this.token) {
							this.sendSeq = first + batch.length - 1;
						}

						return;
					}

					if (response.status < 500) {
						this.reconnect();
						return;
					}
				} catch (err) {
					console.log('http wire send failed', err);
				}

				if (attempt >= sendAttempts) {
					this.reconnect();
					return;
				}

				await sleep(250 * attempt);
			}
		} finally {
			this.sending = false;
			if (this.pending.length > 0) {
				void this.flush();
			}
		}
	}

	// reconnect drops the channel, whose events could not be delivered. The poll gets a new one.
	private reconnect(): void {
		if (this.token === '') {
			return;
		}

		this.token = '';
		this.pending = [];
		this.reconnecting = true;
		this.setConnected(false);
		this.abort?.abort();
	}

	private setConnected(connected: boolean): void {
		if (this.connected === connected) {
			return;
		}

		this.connected = connected;
		if (connected) {
			this.resolveReady?.();
			this.instance.check();
		}

		ConnectionHandler.connectionChanged({ connected });
	}

	private backoff(): number {
		this.retries++;
		return Math.min(2000, 100 * 2 ** (this.retries - 1));
	}
}

function sleep(millis: number): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, millis));
}
