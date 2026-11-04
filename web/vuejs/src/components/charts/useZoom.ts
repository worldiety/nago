/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { type Ref, computed, onBeforeUnmount, onMounted, shallowRef, watch } from 'vue';
import type { Point, Rect } from '@/components/charts/chart';
import type { useChart } from '@/components/charts/useChart';

// zoomDuration is how long a zoom by a double tap, a dragged range, a key or the reset button takes.
const zoomDuration = 350;
// glideDuration is how long the value axis takes to the values a zoom brings into view.
const glideDuration = 250;
// minSpan is the fewest categories a zoom goes down to.
const minSpan = 4;
// wheelNotch is the factor of one notch of a wheel turned with the control or command key held.
const wheelNotch = 1.25;
// slop is how far a finger moves before it drags rather than taps.
const slop = 8;

export type Range = [number, number];

export interface ZoomOptions {
	root: Ref<HTMLElement | undefined>;
	// count returns the number of categories.
	count: () => number;
	// plot returns the area the categories are spread over.
	plot: () => Rect | undefined;
	// inspected is the inspected category, which a drag ends and the keys zoom about.
	inspected: Ref<number>;
	// listeners are those of useChart, which inspect where the zoom leaves the input to them.
	listeners: ReturnType<typeof useChart>['listeners'];
}

interface Press {
	id: number;
	start: Point;
	t: number;
	base: Range;
	// a press becomes a range dragged out with the mouse, a finger dragging the chart or one inspecting it
	mode: 'press' | 'brush' | 'pan' | 'scrub';
}

// useZoom zooms into the category axis, as gift's charts zoom into a continuous one: the mouse drags out the
// range to zoom into, two fingers pinch, a double tap or click zooms in twice where it lands, and the wheel zooms
// with the control or command key held, which is also how a trackpad pinches. A zoomed chart is dragged with a
// finger and scrolled sideways on a trackpad or with the shift key held. The keys + and - zoom about the
// inspected category, the arrow keys keep it in view, and 0 and the reset button show all of it again.
export function useZoom(o: ZoomOptions) {
	// shown is the visible range while the chart is zoomed or on its way back, null while it shows all of it.
	const shown = shallowRef<Range | null>(null);
	const view = computed<Range>(() => shown.value ?? full());
	const zoomable = computed(() => o.count() > minSpan);
	const zoomed = computed(() => shown.value !== null && isZoomed(shown.value));
	// brush is the range the mouse drags out, as positions on the plot.
	const brush = shallowRef<Range | null>(null);
	// keyZoom is where a key zoomed to, for a screen reader to tell.
	const keyZoom = shallowRef<Range | null>(null);

	let anim: { from: Range; to: Range; t0: number; dur: number } | undefined;
	let frame = 0;
	let press: Press | undefined;
	const touches = new Map<number, Point>();
	let pinch: { base: Range; anchor: number; dist: number } | undefined;
	let samples: { t: number; x: number }[] = [];
	let hold = 0;
	let lastTap: { t: number; p: Point } | undefined;
	let lastType = '';

	function full(): Range {
		return [-0.5, Math.max(o.count(), 1) - 0.5];
	}

	function isZoomed(r: Range): boolean {
		const [a, b] = full();
		const eps = (b - a) * 1e-9;
		return r[0] > a + eps || r[1] < b - eps;
	}

	// clamp keeps lo..hi within the categories and at least minSpan wide.
	function clamp(lo: number, hi: number): Range {
		const [a, b] = full();
		const span = Math.min(Math.max(hi - lo, minSpan), b - a);
		lo = Math.min(Math.max((lo + hi) / 2 - span / 2, a), b - span);
		return [lo, lo + span];
	}

	// goal is where the chart is going, or where it is.
	function goal(): Range {
		return anim?.to ?? view.value;
	}

	function settle() {
		if (shown.value && !isZoomed(shown.value)) {
			shown.value = null;
		}
	}

	function stop() {
		anim = undefined;
		cancelAnimationFrame(frame);
		frame = 0;
	}

	// jump shows lo..hi at once.
	function jump(lo: number, hi: number) {
		stop();
		shown.value = clamp(lo, hi);
		settle();
	}

	// animateTo moves to lo..hi in a glide.
	function animateTo(lo: number, hi: number, dur = zoomDuration) {
		anim = { from: view.value, to: clamp(lo, hi), t0: performance.now(), dur };
		frame ||= requestAnimationFrame(step);
	}

	function step(now: number) {
		frame = 0;
		if (!anim) {
			return;
		}

		const u = Math.min(Math.max((now - anim.t0) / anim.dur, 0), 1);
		const e = 1 - Math.pow(1 - u, 3);
		const { from, to } = anim;
		shown.value = [from[0] + (to[0] - from[0]) * e, from[1] + (to[1] - from[1]) * e];
		if (u < 1) {
			frame = requestAnimationFrame(step);
			return;
		}

		anim = undefined;
		settle();
	}

	// xAt returns the position on the category axis under the position x on the plot.
	function xAt(x: number): number {
		const p = o.plot();
		const [lo, hi] = view.value;
		return p ? lo + ((x - p.x) / p.w) * (hi - lo) : lo;
	}

	// zoom scales the visible range by f about the position at on the category axis.
	function zoom(at: number, f: number, animated = true) {
		const [lo, hi] = animated ? goal() : view.value;
		const r: Range = [at - (at - lo) * f, at + (hi - at) * f];
		if (animated) {
			animateTo(...r);
		} else {
			jump(...r);
		}
	}

	// moveBy moves base by dx positions on the plot.
	function moveBy(base: Range, dx: number, dur = 0) {
		const p = o.plot();
		if (!p) {
			return;
		}

		const d = (dx * (base[1] - base[0])) / p.w;
		if (dur > 0) {
			animateTo(base[0] + d, base[1] + d, dur);
		} else {
			jump(base[0] + d, base[1] + d);
		}
	}

	function reset() {
		if (shown.value) {
			animateTo(...full());
		}
	}

	// reveal moves a zoomed chart so that category i is in view.
	function reveal(i: number) {
		if (!shown.value) {
			return;
		}

		const [lo, hi] = goal();
		if (i - 0.5 < lo) {
			animateTo(i - 0.5, i - 0.5 + hi - lo, 150);
		} else if (i + 0.5 > hi) {
			animateTo(i + 0.5 - (hi - lo), i + 0.5, 150);
		}
	}

	// a zoomed chart that showed the end of the data keeps showing it, so that a live series runs on through it
	watch(o.count, (n, old) => {
		const r = shown.value;
		if (!r) {
			return;
		}

		if (n <= minSpan) {
			stop();
			shown.value = null;
			return;
		}

		const d = n > old && r[1] >= old - 0.5 - 1e-9 ? n - old : 0;
		if (anim) {
			anim.from = [anim.from[0] + d, anim.from[1] + d];
			anim.to = clamp(anim.to[0] + d, anim.to[1] + d);
		}

		shown.value = clamp(r[0] + d, r[1] + d);
		settle();
	});

	function local(e: MouseEvent): Point {
		const r = o.root.value!.getBoundingClientRect();
		return { x: e.clientX - r.left, y: e.clientY - r.top };
	}

	function inPlot(p: Point): boolean {
		const r = o.plot();
		return !!r && p.x >= r.x && p.x <= r.x + r.w && p.y >= r.y && p.y <= r.y + r.h;
	}

	function onPointerDown(e: PointerEvent) {
		lastType = e.pointerType;
		if (!zoomable.value) {
			o.listeners.pointerdown(e);
			return;
		}

		const p = local(e);
		if (e.pointerType === 'mouse') {
			if (e.button !== 0) {
				return;
			}

			o.listeners.pointerdown(e);
			if (inPlot(p)) {
				press = { id: e.pointerId, start: p, t: e.timeStamp, base: view.value, mode: 'press' };
				o.root.value?.setPointerCapture(e.pointerId);
			}

			return;
		}

		touches.set(e.pointerId, p);
		if (touches.size === 2) {
			// two fingers pinch about the middle between them
			const [a, b] = [...touches.values()];
			pinch = {
				base: view.value,
				anchor: xAt((a.x + b.x) / 2),
				dist: Math.max(Math.hypot(a.x - b.x, a.y - b.y), 1),
			};
			press = undefined;
			clearTimeout(hold);
			o.inspected.value = -1;
			return;
		}

		if (touches.size > 2) {
			return;
		}

		o.listeners.pointerdown(e);
		press = { id: e.pointerId, start: p, t: e.timeStamp, base: view.value, mode: 'press' };
		// a finger that rests on a zoomed chart inspects as it moves, instead of dragging it
		hold = window.setTimeout(() => {
			if (press?.mode === 'press') {
				press.mode = 'scrub';
			}
		}, 400);
	}

	function onPointerMove(e: PointerEvent) {
		if (!zoomable.value) {
			o.listeners.pointermove(e);
			return;
		}

		const p = local(e);
		if (e.pointerType === 'mouse') {
			if (press?.mode === 'press' && Math.abs(p.x - press.start.x) > 3) {
				press.mode = 'brush';
				o.inspected.value = -1;
			}

			const r = o.plot();
			if (press?.mode === 'brush' && r) {
				brush.value = [press.start.x, Math.min(Math.max(p.x, r.x), r.x + r.w)];
				return;
			}

			o.listeners.pointermove(e);
			return;
		}

		if (!touches.has(e.pointerId)) {
			// a pen above the chart
			o.listeners.pointermove(e);
			return;
		}

		touches.set(e.pointerId, p);
		if (pinch) {
			const r = o.plot();
			if (touches.size < 2 || !r) {
				return;
			}

			const [a, b] = [...touches.values()];
			const span = ((pinch.base[1] - pinch.base[0]) * pinch.dist) / Math.max(Math.hypot(a.x - b.x, a.y - b.y), 1);
			const lo = pinch.anchor - (((a.x + b.x) / 2 - r.x) / r.w) * span;
			jump(lo, lo + span);
			return;
		}

		if (!press) {
			return;
		}

		const dx = p.x - press.start.x;
		const dy = p.y - press.start.y;
		if (press.mode === 'press' && Math.hypot(dx, dy) > slop) {
			clearTimeout(hold);
			if (zoomed.value && Math.abs(dx) > Math.abs(dy)) {
				// a finger drags a zoomed chart sideways, anything else goes on to the page
				press.mode = 'pan';
				samples = [];
				o.inspected.value = -1;
			} else {
				press.mode = 'scrub';
			}
		}

		if (press.mode === 'pan') {
			moveBy(press.base, -dx);
			samples = samples.filter((s) => e.timeStamp - s.t <= 100);
			samples.push({ t: e.timeStamp, x: p.x });
			return;
		}

		if (press.mode === 'scrub' || !zoomed.value) {
			o.listeners.pointermove(e);
		}
	}

	function onPointerUp(e: PointerEvent) {
		if (e.pointerType === 'mouse') {
			const b = brush.value;
			if (press?.mode === 'brush' && b) {
				const x0 = Math.min(b[0], b[1]);
				const x1 = Math.max(b[0], b[1]);
				if (x1 - x0 >= 8) {
					animateTo(xAt(x0), xAt(x1));
				}
			}

			brush.value = null;
			press = undefined;
			return;
		}

		if (!touches.delete(e.pointerId)) {
			return;
		}

		clearTimeout(hold);
		if (pinch) {
			// the finger left after a pinch does nothing until it is lifted as well
			if (touches.size === 0) {
				pinch = undefined;
			}

			return;
		}

		const p = local(e);
		if (press?.mode === 'pan') {
			// the chart glides on to where the flick would have taken it
			const s = samples.find((s) => e.timeStamp - s.t <= 100);
			const v = s && e.timeStamp > s.t ? (p.x - s.x) / (e.timeStamp - s.t) : 0;
			moveBy(press.base, -(p.x - press.start.x + v * 300), 500);
		} else if (
			press &&
			e.timeStamp - press.t < 300 &&
			Math.hypot(p.x - press.start.x, p.y - press.start.y) <= slop
		) {
			// a second tap close to the one before zooms in twice
			if (lastTap && e.timeStamp - lastTap.t < 300 && Math.hypot(p.x - lastTap.p.x, p.y - lastTap.p.y) < 30) {
				zoom(xAt(p.x), 0.5);
				lastTap = undefined;
			} else {
				lastTap = { t: e.timeStamp, p };
			}
		}

		press = undefined;
	}

	function onPointerCancel(e: PointerEvent) {
		touches.delete(e.pointerId);
		clearTimeout(hold);
		if (touches.size === 0) {
			pinch = undefined;
		}

		press = undefined;
		brush.value = null;
	}

	function onDoubleClick(e: MouseEvent) {
		// a double tap of a finger is a pair of taps, not this
		if (!zoomable.value || lastType !== 'mouse') {
			return;
		}

		const p = local(e);
		if (inPlot(p)) {
			zoom(xAt(p.x), 0.5);
		}
	}

	function onWheel(e: WheelEvent) {
		if (!zoomable.value) {
			return;
		}

		const px = e.deltaMode === WheelEvent.DOM_DELTA_PIXEL ? 1 : 40;
		if (e.ctrlKey || e.metaKey) {
			// a mouse wheel turns in notches, a trackpad pinches in small steps
			const dy = e.deltaY * px;
			if (dy === 0) {
				return;
			}

			e.preventDefault();
			const f =
				Number.isInteger(dy) && Math.abs(dy) >= 50
					? Math.pow(wheelNotch, Math.sign(dy))
					: Math.pow(2, dy * 0.01);
			zoom(xAt(local(e).x), f, false);
			reinspect(e);
			return;
		}

		if (!zoomed.value) {
			return;
		}

		const dx = e.shiftKey && e.deltaX === 0 ? e.deltaY : e.deltaX;
		if (dx === 0 || (!e.shiftKey && Math.abs(e.deltaY) >= Math.abs(dx))) {
			return;
		}

		e.preventDefault();
		moveBy(view.value, dx * px);
		reinspect(e);
	}

	// reinspect inspects what moved under the mouse.
	function reinspect(e: WheelEvent) {
		if (o.inspected.value >= 0) {
			o.listeners.pointermove(e);
		}
	}

	function onKeyDown(e: KeyboardEvent) {
		if (!zoomable.value || e.ctrlKey || e.metaKey || e.altKey) {
			o.listeners.keydown(e);
			return;
		}

		const [lo, hi] = goal();
		const at = o.inspected.value >= 0 ? o.inspected.value : (lo + hi) / 2;
		switch (e.key) {
			case '+':
			case '=':
				zoom(at, 0.5);
				break;
			case '-':
			case '_':
				zoom(at, 2);
				break;
			case '0':
				reset();
				break;
			default:
				o.listeners.keydown(e);
				if (o.inspected.value >= 0) {
					reveal(o.inspected.value);
				}

				return;
		}

		e.preventDefault();
		keyZoom.value = goal();
	}

	onMounted(() => o.root.value?.addEventListener('wheel', onWheel, { passive: false }));

	onBeforeUnmount(() => {
		o.root.value?.removeEventListener('wheel', onWheel);
		stop();
		clearTimeout(hold);
	});

	return {
		view,
		zoomable,
		zoomed,
		// active is true while the chart is zoomed or on its way back.
		active: computed(() => shown.value !== null),
		brush,
		keyZoom,
		reset,
		listeners: {
			...o.listeners,
			pointerdown: onPointerDown,
			pointermove: onPointerMove,
			pointerup: onPointerUp,
			pointercancel: onPointerCancel,
			dblclick: onDoubleClick,
			keydown: onKeyDown,
		},
	};
}

// useGlide follows the range target: in a glide while active says so, as the value axis moves to the values a
// zoom brings into view, and at once otherwise.
export function useGlide(target: () => Range | undefined, active: () => boolean) {
	const shown = shallowRef<Range | undefined>();
	let anim: { from: Range; to: Range; t0: number } | undefined;
	let frame = 0;

	function step(now: number) {
		frame = 0;
		if (!anim) {
			return;
		}

		const u = Math.min(Math.max((now - anim.t0) / glideDuration, 0), 1);
		const e = 1 - Math.pow(1 - u, 3);
		const { from, to } = anim;
		shown.value = [from[0] + (to[0] - from[0]) * e, from[1] + (to[1] - from[1]) * e];
		if (u < 1) {
			frame = requestAnimationFrame(step);
		} else {
			anim = undefined;
		}
	}

	watch(
		target,
		(to) => {
			const from = shown.value;
			if (!to || !from || !(active() || anim)) {
				anim = undefined;
				shown.value = to;
				return;
			}

			const going = anim?.to ?? from;
			if (to[0] === going[0] && to[1] === going[1]) {
				return;
			}

			anim = { from, to, t0: performance.now() };
			frame ||= requestAnimationFrame(step);
		},
		{ immediate: true }
	);

	onBeforeUnmount(() => cancelAnimationFrame(frame));

	return shown;
}
