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
import { TextFieldSync } from '../src/components/shared/textFieldSync.ts';

describe('TextFieldSync', () => {
	test('a rendering does not overwrite input, which has not been sent yet', () => {
		// the user types "1", the debounce runs, and the server renders on its own with the old value
		const sync = new TextFieldSync('');
		assert.equal(sync.hasUnsentInput('1'), true);
		assert.equal(sync.adopt('1', '', 0), false);
		// the pending debounce must still transmit it
		assert.equal(sync.hasUnsentInput('1'), true);
	});

	test('a rendering with the sent value does not overwrite what has been typed afterwards', () => {
		// "1" has been sent, the user has typed "2" meanwhile and a rendering still carries "1"
		const sync = new TextFieldSync('');
		sync.sent('1', 4);
		assert.equal(sync.adopt('12', '1', 4), false);
		assert.equal(sync.hasUnsentInput('12'), true);
	});

	test('a server side reset of the sent value is adopted while the field keeps the focus', () => {
		// chat input: enter sends "hello" and the submit handler empties the state
		const sync = new TextFieldSync('');
		sync.sent('hello', 7);
		assert.equal(sync.hasUnsentInput('hello'), false);
		assert.equal(sync.adopt('hello', '', 7), true);
		assert.equal(sync.hasUnsentInput(''), false);
	});

	test('a server side change is adopted without local edits', () => {
		const sync = new TextFieldSync('a');
		assert.equal(sync.adopt('a', 'b', 0), true);
		assert.equal(sync.hasUnsentInput('b'), false);
		assert.equal(sync.hasUnsentInput('a'), true);
	});

	test('a rendering with the known value leaves the input alone', () => {
		const sync = new TextFieldSync('');
		sync.sent('1', 4);
		assert.equal(sync.adopt('1', '1', 0), false);
		assert.equal(sync.hasUnsentInput('1'), false);
	});

	test('deleting what has just been sent is unsent input, even if the server still shows the old value', () => {
		// "1" is on its way, the rendering still shows "" and the user deletes the "1" again
		const sync = new TextFieldSync('');
		sync.sent('1', 4);
		assert.equal(sync.hasUnsentInput(''), true);
		assert.equal(sync.adopt('', '1', 4), false);
	});

	test('a rendering made before the server received the sent value does not throw the field back', () => {
		// "1" is on its way with rid 5, the server renders on its own (rid 0) before it receives "1"
		const sync = new TextFieldSync('');
		sync.sent('1', 5);
		assert.equal(sync.adopt('1', '', 3), false);
		// the answer to 5 carries "1", a later server side change is adopted again
		assert.equal(sync.adopt('1', '1', 5), false);
		assert.equal(sync.adopt('1', 'x', 5), true);
	});

	test('a server side reset is adopted once the rendering answers the sent request or a later one', () => {
		const sync = new TextFieldSync('');
		sync.sent('hello', 7);
		assert.equal(sync.adopt('hello', 'hel', 6), false); // made before the server received "hello"
		assert.equal(sync.adopt('hello', '', 9), true); // answers 7 or later, the submit handler emptied it
	});

	test('a rendering carrying the sent value ends the flight without an answer of its own', () => {
		// e.g. enter without a change, which the server does not answer with a rendering of its rid
		const sync = new TextFieldSync('a');
		sync.sent('a', 9);
		assert.equal(sync.adopt('a', 'a', 3), false);
		assert.equal(sync.adopt('a', 'b', 3), true);
	});
});
