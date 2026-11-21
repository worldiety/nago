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
import {
	type ElementLike,
	handlesKeysItself,
	isForwardedLabelClick,
} from '../src/components/shared/stackActivation.ts';

/** FakeElement mimics just enough of the DOM: a tag, a parent and for labels the associated control. */
class FakeElement implements ElementLike {
	readonly tag: string;
	readonly parent: FakeElement | null;
	control?: FakeElement | null;
	disabled?: boolean;

	constructor(tag: string, parent: FakeElement | null = null) {
		this.tag = tag;
		this.parent = parent;
	}

	closest(selectors: string): FakeElement | null {
		const tags = selectors.split(',').map((s) => s.trim().replace(/\[.*$/, ''));
		for (let e: FakeElement | null = this; e; e = e.parent) {
			if (tags.includes(e.tag)) return e;
		}
		return null;
	}

	contains(other: ElementLike | null): boolean {
		for (let e = other as FakeElement | null; e; e = e.parent) {
			if (e === this) return true;
		}
		return false;
	}
}

/** card builds a clickable stack with a radio button and a label for it, like a selectable card. */
function card() {
	const page = new FakeElement('div');
	const stack = new FakeElement('div', page);
	const radioContainer = new FakeElement('div', stack);
	const input = new FakeElement('input', radioContainer);
	const label = new FakeElement('label', stack);
	label.control = input;
	const text = new FakeElement('span', stack);
	return { page, stack, input, label, text };
}

describe('isForwardedLabelClick', () => {
	test('a click on the label of a control within the stack is forwarded', () => {
		const { stack, label } = card();
		assert.equal(isForwardedLabelClick(label, stack), true);
	});

	test('a click on other content of the stack is not forwarded', () => {
		const { stack, text } = card();
		assert.equal(isForwardedLabelClick(text, stack), false);
	});

	test('a click on the control itself is not forwarded', () => {
		const { stack, input } = card();
		assert.equal(isForwardedLabelClick(input, stack), false);
	});

	test('a click on the control nested in its label is not forwarded', () => {
		const stack = new FakeElement('div');
		const label = new FakeElement('label', stack);
		const input = new FakeElement('input', label);
		label.control = input;
		assert.equal(isForwardedLabelClick(input, stack), false);
		assert.equal(isForwardedLabelClick(label, stack), true);
	});

	test('the browser does not forward the click on the label of a disabled control', () => {
		const { stack, input, label } = card();
		input.disabled = true;
		assert.equal(isForwardedLabelClick(label, stack), false);
	});

	test('the click on a label for a control outside the stack is handled by the stack', () => {
		const { page, stack, label } = card();
		label.control = new FakeElement('input', page);
		assert.equal(isForwardedLabelClick(label, stack), false);
	});

	test('a label without a control is handled by the stack', () => {
		const { stack, label } = card();
		label.control = null;
		assert.equal(isForwardedLabelClick(label, stack), false);
	});
});

describe('handlesKeysItself', () => {
	test('a form control within the stack handles its keys', () => {
		const { stack, input } = card();
		assert.equal(handlesKeysItself(input, stack), true);
	});

	test('the focused stack itself does not', () => {
		const { stack } = card();
		assert.equal(handlesKeysItself(stack, stack), false);
	});

	test('a stack rendered as button is not a control within itself', () => {
		const stack = new FakeElement('button');
		const text = new FakeElement('span', stack);
		assert.equal(handlesKeysItself(text, stack), false);
	});

	test('a control around the stack is not within the stack', () => {
		const button = new FakeElement('button');
		const stack = new FakeElement('div', button);
		assert.equal(handlesKeysItself(stack, stack), false);
	});
});
