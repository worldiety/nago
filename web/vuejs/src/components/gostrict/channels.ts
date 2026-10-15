/*
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import type { GostrictMessage } from '@/shared/proto/nprotoc_gen';

/**
 * The channels connect the gostrict programs of the components with their backends by the id of the component.
 * The backend opens a channel with GostrictListen, usually before the component is mounted, and may deliver
 * messages before the component has loaded its program, so a channel keeps them until then.
 */

type Receiver = (msgs: GostrictMessage[]) => void;
type Sender = (msgs: GostrictMessage[]) => void;

interface Entry {
	send?: Sender;
	receiver?: Receiver;
	pending: GostrictMessage[];
}

// maxPending limits the messages, which a channel keeps for a component, which is not mounted
const maxPending = 10000;

const entries = new Map<string, Entry>();

function entry(id: string): Entry {
	let e = entries.get(id);
	if (!e) {
		e = { pending: [] };
		entries.set(id, e);
	}
	return e;
}

/**
 * openChannel registers the sender of the messages of the program to the backend. A window, which replaces the
 * former one, opens the channel again and drops the messages, which the former window had sent.
 */
export function openChannel(id: string, send: Sender) {
	const e = entry(id);
	e.send = send;
	if (!e.receiver) e.pending = [];
}

/**
 * deliver passes the messages of the backend to the program or keeps them, until the program is attached.
 */
export function deliver(id: string, msgs: GostrictMessage[]) {
	const e = entry(id);
	if (e.receiver) {
		e.receiver(msgs);
		return;
	}
	e.pending.push(...msgs);
	if (e.pending.length > maxPending) {
		console.warn('gostrict: dropping messages for a component, which is not mounted', id);
		e.pending.splice(0, e.pending.length - maxPending);
	}
}

export interface ProgramChannel {
	// send passes the messages of the program to the backend
	send(msgs: GostrictMessage[]): void;
	detach(): void;
}

/**
 * attach connects the program of the component with the channel and passes the kept messages to it.
 */
export function attach(id: string, receiver: Receiver): ProgramChannel {
	const e = entry(id);
	e.receiver = receiver;
	if (e.pending.length > 0) {
		const pending = e.pending;
		e.pending = [];
		receiver(pending);
	}
	return {
		send(msgs: GostrictMessage[]) {
			if (msgs.length === 0) return;
			const cur = entries.get(id);
			if (!cur?.send) {
				console.debug('gostrict: no backend receives the messages of', id);
				return;
			}
			cur.send(msgs);
		},
		detach() {
			const cur = entries.get(id);
			if (cur && cur.receiver === receiver) cur.receiver = undefined;
		},
	};
}
