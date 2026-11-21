/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

/**
 * The helpers decide, whether a clickable stack sends its action for an event, which has been dispatched to one of its
 * descendants. They prevent, that a single gesture causes two server events: one by the stack and one by a form control
 * within it, e.g. a radio button within a selectable card. Then, the server renders for the first event and the
 * frontend discards that rendering, because it already waits for the answer of the second one (see App.vue). If the
 * second event does not change anything, the server does not render for it at all and the frontend keeps showing the
 * outdated tree.
 *
 * The functions only use a few members of the DOM types, so that node can test them with plain objects, see
 * test/stackActivation.test.ts.
 */

/** ElementLike is the part of an Element, which the helpers use. */
export interface ElementLike {
	closest(selectors: string): ElementLike | null;
	contains(other: ElementLike | null): boolean;
	// only for labels, see HTMLLabelElement.control
	readonly control?: (ElementLike & { readonly disabled?: boolean }) | null;
}

/**
 * isForwardedLabelClick tells, whether the click hits a label, whose enabled control is inside the stack as well.
 * The browser forwards such a click as a second, synthetic click to the control, which bubbles up through the stack
 * again. The stack handles that one, unless the control handles it itself (e.g. a radio button or a checkbox). Without
 * this check, the stack sends its action for the click on the label and the control sends its state for the
 * forwarded click.
 */
export function isForwardedLabelClick(target: ElementLike | null, stack: ElementLike): boolean {
	const label = target?.closest('label');
	if (!label || !stack.contains(label)) {
		return false;
	}

	const control = label.control;
	if (!control || control.disabled || !stack.contains(control)) {
		// the browser does not forward the click, so the stack handles it as before
		return false;
	}

	// a click on the control itself is not forwarded, even if the control is nested in its label
	return !control.contains(target);
}

/**
 * handlesKeysItself tells, whether the key event targets an element within the stack, which handles the enter and
 * space key itself, e.g. a radio button, which is selected by space. The stack must not send its action for these keys
 * then.
 */
export function handlesKeysItself(target: ElementLike | null, stack: ElementLike): boolean {
	const control = target?.closest(
		'input, textarea, select, button, a[href], [contenteditable]:not([contenteditable="false"])'
	);
	return !!control && control !== stack && stack.contains(control);
}
