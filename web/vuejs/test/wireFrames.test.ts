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
import { describe, test } from 'node:test';
import { decodeFrames, encodeFrames } from '../src/shared/network/wireFrames.ts';
import {
	preferHttp,
	rememberHttp,
	rememberHttpMillis,
	withoutTransportParam,
} from '../src/shared/network/wireTransport.ts';

describe('wire frames', () => {
	test('match the encoding of the backend', () => {
		const big = new Uint8Array(200).fill(7);
		const buf = encodeFrames(300, [new Uint8Array([1, 2, 3]), new Uint8Array(0), big]);
		// like poll.WriteFrames: uvarint 300, then each message with its uvarint length
		const want = [0xac, 0x02, 0x03, 1, 2, 3, 0x00, 0xc8, 0x01, ...big];
		assert.deepEqual(Array.from(buf), want);

		const { first, msgs } = decodeFrames(buf);
		assert.equal(first, 300);
		assert.deepEqual(
			msgs.map((m) => Array.from(m)),
			[[1, 2, 3], [], Array.from(big)]
		);
	});

	test('a body without messages holds the number of the next one', () => {
		const { first, msgs } = decodeFrames(encodeFrames(2 ** 53 - 1, []));
		assert.equal(first, 2 ** 53 - 1);
		assert.equal(msgs.length, 0);
	});

	test('a truncated body fails', () => {
		const buf = encodeFrames(1, [new Uint8Array(10)]);
		assert.throws(() => decodeFrames(buf.subarray(0, buf.length - 1)));
		assert.throws(() => decodeFrames(new Uint8Array([0x80])));
		assert.throws(() => decodeFrames(new Uint8Array(0)));
	});
});

class MemoryStorage {
	private readonly items = new Map<string, string>();

	getItem(key: string): string | null {
		return this.items.get(key) ?? null;
	}

	setItem(key: string, value: string): void {
		this.items.set(key, value);
	}
}

describe('wire transport', () => {
	test('the query of the page or a recent fallback prefer http', () => {
		const storage = new MemoryStorage() as unknown as Storage;
		assert.equal(preferHttp('', storage, 1000), false);
		assert.equal(preferHttp('?a=1&_transport=http', storage, 1000), true);
		assert.equal(preferHttp('?_transport=ws', storage, 1000), false);

		rememberHttp(storage, 1000);
		assert.equal(preferHttp('', storage, 1000 + rememberHttpMillis - 1), true);
		assert.equal(preferHttp('', storage, 1000 + rememberHttpMillis), false);
		assert.equal(preferHttp('', null, 1000), false);
	});

	test('the parameter is not passed to the application', () => {
		assert.equal(withoutTransportParam('a=1&_transport=http&b=x%20y'), 'a=1&b=x+y');
		assert.equal(withoutTransportParam('a=1&b=x%20y'), 'a=1&b=x%20y');
		assert.equal(withoutTransportParam(''), '');
	});
});
