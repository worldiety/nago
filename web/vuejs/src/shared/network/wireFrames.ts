/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

/**
 * Frames is the body of the requests of the http wire in both directions, see poll.WriteFrames of the backend: the
 * number of the first message, followed by each message with its length. A message is what a websocket message
 * holds, thus an encoded nago event.
 */
export interface Frames {
	first: number;
	msgs: Uint8Array[];
}

export function encodeFrames(first: number, msgs: Uint8Array[]): Uint8Array {
	const parts: number[] = [];
	appendUvarint(parts, first);
	let size = 0;
	const lengths = msgs.map((msg) => {
		const prefix: number[] = [];
		appendUvarint(prefix, msg.length);
		size += prefix.length + msg.length;
		return prefix;
	});

	const buf = new Uint8Array(parts.length + size);
	buf.set(parts, 0);
	let offset = parts.length;
	msgs.forEach((msg, i) => {
		buf.set(lengths[i], offset);
		offset += lengths[i].length;
		buf.set(msg, offset);
		offset += msg.length;
	});

	return buf;
}

export function decodeFrames(buf: Uint8Array): Frames {
	const reader = { buf, offset: 0 };
	const first = readUvarint(reader);
	const msgs: Uint8Array[] = [];
	while (reader.offset < buf.length) {
		const length = readUvarint(reader);
		if (reader.offset + length > buf.length) {
			throw new Error(`frame ${msgs.length} exceeds the body`);
		}

		msgs.push(buf.subarray(reader.offset, reader.offset + length));
		reader.offset += length;
	}

	return { first, msgs };
}

// appendUvarint appends the value like Go's binary.AppendUvarint. It is exact for values up to 2^53-1.
function appendUvarint(out: number[], value: number): void {
	let v = value;
	while (v >= 0x80) {
		out.push((v % 0x80) | 0x80);
		v = Math.floor(v / 0x80);
	}

	out.push(v);
}

function readUvarint(reader: { buf: Uint8Array; offset: number }): number {
	let value = 0;
	let scale = 1;
	for (let i = 0; i < 10; i++) {
		if (reader.offset >= reader.buf.length) {
			throw new Error('incomplete uvarint');
		}

		const b = reader.buf[reader.offset++];
		value += (b & 0x7f) * scale;
		if (b < 0x80) {
			return value;
		}

		scale *= 0x80;
	}

	throw new Error('uvarint overflows');
}
