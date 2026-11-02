/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { onBeforeUnmount, onMounted, ref, shallowRef } from 'vue';
import { type Point, TextMeasurer } from '@/components/charts/chart';

// useChart follows the size of the chart's element and the inspection: the category or the slice under a mouse,
// a finger or a pen, or the one chosen with the arrow keys. count returns how many there are and hit which one is
// at a position, or -1.
export function useChart(count: () => number, hit: (x: number, y: number) => number) {
	const root = ref<HTMLElement>();
	const width = ref(0);
	const height = ref(0);
	const measure = shallowRef(new TextMeasurer('sans-serif'));
	// inspected is the inspected category or slice, -1 if there is none.
	const inspected = ref(-1);
	// pointer is where the pointer inspects, null if the keyboard does.
	const pointer = ref<Point | null>(null);
	const keyboard = ref(false);
	let observer: ResizeObserver | undefined;

	function resize() {
		if (!root.value) {
			return;
		}

		width.value = root.value.clientWidth;
		height.value = root.value.clientHeight;
	}

	function local(e: PointerEvent): Point {
		const r = root.value!.getBoundingClientRect();
		return { x: e.clientX - r.left, y: e.clientY - r.top };
	}

	function inspectAt(e: PointerEvent) {
		const p = local(e);
		inspected.value = hit(p.x, p.y);
		pointer.value = p;
		keyboard.value = false;
	}

	function onPointerLeave(e: PointerEvent) {
		// a finger leaves when it is lifted, the inspection stays until the next touch outside
		if (e.pointerType === 'mouse') {
			inspected.value = -1;
		}
	}

	function onDocumentPointerDown(e: PointerEvent) {
		if (root.value && !root.value.contains(e.target as Node)) {
			inspected.value = -1;
		}
	}

	function onKeyDown(e: KeyboardEvent) {
		const n = count();
		if (n === 0) {
			return;
		}

		let i = inspected.value;
		switch (e.key) {
			case 'ArrowRight':
			case 'ArrowDown':
				i = i < 0 ? 0 : Math.min(i + 1, n - 1);
				break;
			case 'ArrowLeft':
			case 'ArrowUp':
				i = i < 0 ? n - 1 : Math.max(i - 1, 0);
				break;
			case 'Home':
				i = 0;
				break;
			case 'End':
				i = n - 1;
				break;
			case 'Escape':
				if (i < 0) {
					return;
				}

				i = -1;
				break;
			default:
				return;
		}

		e.preventDefault();
		inspected.value = i;
		pointer.value = null;
		keyboard.value = true;
	}

	function onFocusOut(e: FocusEvent) {
		if (keyboard.value && !root.value?.contains(e.relatedTarget as Node)) {
			inspected.value = -1;
		}
	}

	onMounted(() => {
		const el = root.value!;
		measure.value = new TextMeasurer(getComputedStyle(el).fontFamily || 'sans-serif');
		resize();
		observer = new ResizeObserver(resize);
		observer.observe(el);
		document.addEventListener('pointerdown', onDocumentPointerDown, true);
	});

	onBeforeUnmount(() => {
		observer?.disconnect();
		document.removeEventListener('pointerdown', onDocumentPointerDown, true);
	});

	return {
		root,
		width,
		height,
		measure,
		inspected,
		pointer,
		keyboard,
		listeners: {
			pointermove: inspectAt,
			pointerdown: inspectAt,
			pointerleave: onPointerLeave,
			keydown: onKeyDown,
			focusout: onFocusOut,
		},
	};
}
