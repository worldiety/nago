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
import { OptimisticToggle } from '../src/components/shared/optimisticToggle.ts';

const state = 7;

describe('OptimisticToggle', () => {
	test('a toggle shows at once', () => {
		const sync = new OptimisticToggle(false, state);
		sync.toggle(true, 4);
		assert.equal(sync.value, true);
	});

	test('the answer to the toggle confirms it', () => {
		const sync = new OptimisticToggle(false, state);
		sync.toggle(true, 4);
		assert.equal(sync.adopt(true, state, 4), true);
	});

	test('a rendering made before the server received the toggle keeps the local value', () => {
		// e.g. a rendering the server made on its own (request tracing id 0) or an answer to an older request
		const sync = new OptimisticToggle(false, state);
		sync.toggle(true, 4);
		assert.equal(sync.adopt(false, state, 3), true);
		assert.equal(sync.adopt(false, state, 3), true);
		// and the answer still confirms it afterwards
		assert.equal(sync.adopt(true, state, 4), true);
	});

	test('an answer with the unchanged value falls back, because the server did not take the toggle', () => {
		// e.g. the state has been removed from the tree in the meantime and the server rendered the stale tree again
		const sync = new OptimisticToggle(false, state);
		sync.toggle(true, 4);
		assert.equal(sync.adopt(false, state, 4), false);
	});

	test('a later rendering, which answers another request, decides as well', () => {
		const sync = new OptimisticToggle(false, state);
		sync.toggle(true, 4);
		assert.equal(sync.adopt(false, state, 9), false);
	});

	test('fast toggles keep the last one until its answer', () => {
		const sync = new OptimisticToggle(false, state);
		sync.toggle(true, 4);
		sync.toggle(false, 5);
		// the answer to the first toggle (App.vue discards it already, but even if not)
		assert.equal(sync.adopt(true, state, 4), false);
		assert.equal(sync.adopt(false, state, 5), false);
	});

	test('without a toggle in flight, the server value is adopted', () => {
		// e.g. the server closes the accordion, because another one has been opened
		const sync = new OptimisticToggle(true, state);
		assert.equal(sync.adopt(false, state, 0), false);
	});

	test('an instance reused for another state starts over with its server value', () => {
		const sync = new OptimisticToggle(false, state);
		sync.toggle(true, 4);
		assert.equal(sync.adopt(false, state + 1, 0), false);
		assert.equal(sync.value, false);
	});
});
