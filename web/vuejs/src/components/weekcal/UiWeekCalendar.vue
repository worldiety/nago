<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div
		ref="root"
		class="weekcal"
		:class="{ disabled: !!ui.disabled, pending: !!pending }"
		:data-weekcal="ui.id ?? ''"
		role="application"
		:aria-label="t('weekcal.label')"
		:aria-disabled="ui.disabled ? 'true' : undefined"
		:style="rootStyle"
		@keydown="onKeydown"
		@focusin="onFocusIn"
		@focusout="onFocusOut"
		@contextmenu="onContextMenu"
		@dragstart.prevent
	>
		<!-- the sizer asks a container, which follows its content, for the width of all days, but, as a replaced
		     element with a relative maximum, never forces it wider than its place -->
		<svg class="sizer" aria-hidden="true" :width="minWidth" height="1"></svg>
		<div v-if="ui.message" class="message" role="alert" data-weekcal-message>{{ ui.message }}</div>

		<div class="grid" :style="gridStyle">
			<div class="corner"></div>
			<div
				v-for="(d, i) in days"
				:key="'head' + i"
				class="head"
				:data-weekcal-day="i"
				:aria-current="d.today ? 'date' : undefined"
				:data-muted="d.muted ? '' : undefined"
				:style="{ gridColumn: String(i + 2) }"
			>
				<div class="title">{{ d.title }}</div>
				<div v-if="d.subtitle" class="subtitle">{{ d.subtitle }}</div>
			</div>

			<div class="axis" :style="{ height: bodyHeight + 'px' }">
				<span
					v-for="h in hours"
					:key="h"
					class="hour"
					:class="{ first: h === range.from, last: h === range.to }"
					:data-weekcal-hour="h"
					:style="{ top: y(h) + 'px' }"
					>{{ hhmm(h) }}</span
				>
			</div>

			<div
				ref="body"
				class="body"
				:style="bodyStyle"
				@pointerdown="onPointerDown"
				@pointermove="onPointerMove"
				@pointerup="onPointerUp"
				@pointercancel="onPointerCancel"
			>
				<div
					v-for="(d, i) in days"
					:key="'col' + i"
					class="col"
					:class="{ today: d.today, muted: d.muted }"
					:style="columnStyle(i)"
				></div>

				<div
					ref="cursorEl"
					class="cursor"
					role="button"
					tabindex="0"
					data-weekcal-cursor
					:data-day="cursor.day"
					:data-minute="cursor.minute"
					:aria-label="dayTitle(cursor.day) + ', ' + hhmm(cursor.minute)"
					:style="boxStyle(cursor.day, cursor.minute, cursor.minute + step, 0, 1)"
				></div>

				<div
					v-for="b in blocks"
					:key="b.id"
					class="event"
					:class="{ movable: b.movable, clickable: canClick }"
					role="button"
					tabindex="0"
					:data-weekcal-event="b.id"
					:data-clipped="b.clipped || undefined"
					:data-fixed="b.fixed ? '' : undefined"
					:data-dragging="b.id === draggingId ? '' : undefined"
					:aria-label="eventLabel(b)"
					:style="eventStyle(b)"
				>
					<div class="event-title">{{ b.title }}</div>
					<div class="event-time">{{ hhmm(b.from) }}–{{ hhmm(b.to) }}</div>
					<div v-if="b.text" class="event-text">{{ b.text }}</div>
					<div v-if="b.movable && canMove && !ui.disabled" class="resize" :style="resizeStyle(b)"></div>
				</div>

				<template v-for="(hc, i) in hidden" :key="'hidden' + i">
					<span
						v-if="hc.before > 0"
						class="hidden-count before"
						data-weekcal-hidden="before"
						role="img"
						:aria-label="t('weekcal.hiddenBefore', { n: hc.before })"
						:style="hiddenStyle(i, 'before')"
						>{{ hc.before }}</span
					>
					<span
						v-if="hc.after > 0"
						class="hidden-count after"
						data-weekcal-hidden="after"
						role="img"
						:aria-label="t('weekcal.hiddenAfter', { n: hc.after })"
						:style="hiddenStyle(i, 'after')"
						>{{ hc.after }}</span
					>
				</template>

				<div
					v-if="shownPreview"
					class="preview"
					data-weekcal-preview
					:data-day="shownPreview.day"
					:data-from="shownPreview.from"
					:data-to="shownPreview.to"
					:style="boxStyle(shownPreview.day, shownPreview.from, shownPreview.to, 0, 1)"
				>
					<span class="preview-time">{{ hhmm(shownPreview.from) }}–{{ hhmm(shownPreview.to) }}</span>
				</div>

				<div v-if="nowLine" class="now" data-weekcal-now :style="nowLine"></div>
			</div>
		</div>

		<div class="sr-only" aria-live="polite">{{ announcement }}</div>
	</div>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, onBeforeUpdate, onMounted, onUpdated, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import ConnectionHandler from '@/shared/network/connectionHandler';
import type { ConnectionState } from '@/shared/network/connectionState';
import { UpdateStateValueRequested, WeekCalendar } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: WeekCalendar;
}>();

const { t } = useI18n();
const serviceAdapter = useServiceAdapter();

const root = ref<HTMLElement>();
const body = ref<HTMLElement>();
const cursorEl = ref<HTMLElement>();

// the gestures, which wait for a long press of a finger, and their thresholds (L-N7-3)
const MOUSE_DRAG_PX = 4;
const TOUCH_HOLD_MS = 500;
const TOUCH_SLOP_PX = 10;
// the lower edge of a block, which changes its duration: 6px, at most a third of the block
const RESIZE_PX = 6;

// --- model ---

interface Interval {
	day: number;
	from: number;
	to: number;
}

interface Block extends Interval {
	id: string;
	title: string;
	text: string;
	color: string;
	fixed: boolean;
	// movable tells, whether the block lies entirely in the range and is not fixed
	movable: boolean;
	clipped: string;
	// the shown part within the range and its lane among the overlapping blocks of its day
	top: number;
	bottom: number;
	lane: number;
	lanes: number;
}

const days = computed(() =>
	(props.ui.days?.value ?? []).map((d) => ({
		title: d.title ?? '',
		subtitle: d.subtitle ?? '',
		today: !!d.today,
		muted: !!d.muted,
	}))
);
const n = computed(() => Math.max(days.value.length, 1));
const range = computed(() => {
	const from = Number(props.ui.from ?? 0);
	const to = Number(props.ui.to ?? 0);
	return to > from ? { from, to } : { from: 360, to: 1320 };
});
const step = computed(() => Number(props.ui.step ?? 0) || 15);
const clickSpan = computed(() => Number(props.ui.clickSpan ?? 0) || 60);
const hourHeight = computed(() => Number(props.ui.hourHeight ?? 0) || 48);
const bodyHeight = computed(() => ((range.value.to - range.value.from) / 60) * hourHeight.value);

const canClick = computed(() => !!props.ui.canClick && !!props.ui.input);
const canSelect = computed(() => !!props.ui.canSelect && !!props.ui.input);
const canMove = computed(() => !!props.ui.canMove && !!props.ui.input);

const events = computed(() =>
	(props.ui.events?.value ?? []).map((e) => ({
		id: e.id ?? '',
		day: Number(e.day ?? 0),
		from: Number(e.from ?? 0),
		to: Number(e.to ?? 0),
		title: e.title ?? '',
		text: e.text ?? '',
		color: e.color ?? '',
		fixed: !!e.fixed,
	}))
);

// blocks are the events, which lie at least partly within the range, in the order of the tab key: day, beginning
const blocks = computed<Block[]>(() => {
	const r = range.value;
	const res: Block[] = [];
	for (const e of events.value) {
		if (e.to <= r.from || e.from >= r.to) continue;
		const clipped = [e.from < r.from ? 'top' : '', e.to > r.to ? 'bottom' : ''].filter((x) => x).join(' ');
		res.push({
			...e,
			movable: !e.fixed && e.from >= r.from && e.to <= r.to,
			clipped,
			top: Math.max(e.from, r.from),
			bottom: Math.min(e.to, r.to),
			lane: 0,
			lanes: 1,
		});
	}
	res.sort((a, b) => a.day - b.day || a.from - b.from || a.to - b.to || (a.id < b.id ? -1 : a.id > b.id ? 1 : 0));

	// overlapping blocks of a day share its width: a cluster of transitively overlapping blocks gets as many lanes
	// as it needs at most at once
	for (let i = 0; i < res.length;) {
		let j = i;
		let end = res[i].bottom;
		const laneEnds: number[] = [];
		const cluster: Block[] = [];
		while (j < res.length && res[j].day === res[i].day && (j === i || res[j].top < end)) {
			const b = res[j];
			let lane = laneEnds.findIndex((le) => le <= b.top);
			if (lane < 0) {
				lane = laneEnds.length;
				laneEnds.push(b.bottom);
			} else {
				laneEnds[lane] = b.bottom;
			}
			b.lane = lane;
			cluster.push(b);
			end = Math.max(end, b.bottom);
			j++;
		}
		for (const b of cluster) b.lanes = laneEnds.length;
		i = j;
	}
	return res;
});

// hidden counts the events per day, which lie entirely before or after the range
const hidden = computed(() => {
	const res = days.value.map(() => ({ before: 0, after: 0 }));
	const r = range.value;
	for (const e of events.value) {
		if (e.day < 0 || e.day >= res.length) continue;
		if (e.to <= r.from) res[e.day].before++;
		else if (e.from >= r.to) res[e.day].after++;
	}
	return res;
});

const hours = computed(() => {
	const res: number[] = [];
	for (let h = Math.ceil(range.value.from / 60) * 60; h <= range.value.to; h += 60) res.push(h);
	return res;
});

const nowLine = computed(() => {
	if (!props.ui.showNow) return null;
	const day = Number(props.ui.nowDay ?? 0);
	const minute = Number(props.ui.nowMinute ?? 0);
	if (day < 0 || day >= n.value || minute < range.value.from || minute > range.value.to) return null;
	return { left: pct(day), width: pct(1), top: y(minute) + 'px' };
});

// --- geometry ---

const AXIS_WIDTH = 52;
const MIN_DAY_WIDTH = 96;
const rootStyle = computed(() => frameCSS(props.ui.frame).join(';'));
// The calendar is at most as wide as its place and scrolls itself: a day is at least 96dp wide, the axis stays. Only
// the width of the grid, a maximum of the place and all days, keeps the days; its percentage leaves the days out of
// the smallest width, which the calendar asks of a container, so a container, which follows its content, does not
// widen beyond its place. The sizer asks for the width of all days as the preferred one.
const minWidth = computed(() => AXIS_WIDTH + MIN_DAY_WIDTH * n.value);
const gridStyle = computed(
	() =>
		`grid-template-columns:${AXIS_WIDTH}px repeat(${n.value}, minmax(0, 1fr));` +
		`grid-template-rows:auto ${bodyHeight.value}px;width:max(100%, ${minWidth.value}px)`
);
const bodyStyle = computed(() => {
	const offset = ((Math.ceil(range.value.from / 60) * 60 - range.value.from) / 60) * hourHeight.value;
	return {
		gridColumn: `2 / span ${n.value}`,
		height: bodyHeight.value + 'px',
		backgroundSize: `100% ${hourHeight.value}px`,
		backgroundPositionY: offset + 'px',
	};
});

function pct(days: number): string {
	return (days / n.value) * 100 + '%';
}

function y(minute: number): number {
	return ((minute - range.value.from) / 60) * hourHeight.value;
}

function columnStyle(day: number) {
	return { left: pct(day), width: pct(1) };
}

function boxStyle(day: number, from: number, to: number, lane: number, lanes: number) {
	const top = Math.max(from, range.value.from);
	const bottom = Math.min(to, range.value.to);
	return {
		left: `calc(${pct(day)} + ${lane} * ${pct(1)} / ${lanes})`,
		width: `calc(${pct(1)} / ${lanes})`,
		top: y(top) + 'px',
		height: Math.max(y(bottom) - y(top), 0) + 'px',
	};
}

function eventStyle(b: Block) {
	const s: Record<string, string> = boxStyle(b.day, b.top, b.bottom, b.lane, b.lanes);
	if (b.color) {
		const c = colorValue(b.color);
		s['--weekcal-event'] = c;
	}
	return s;
}

function resizeStyle(b: Block) {
	const h = y(b.bottom) - y(b.top);
	return { height: Math.min(RESIZE_PX, h / 3) + 'px' };
}

function hiddenStyle(day: number, side: 'before' | 'after') {
	// at the right edge of the column, so that it does not cover the title of a block, which begins at the top
	const right = `calc(${pct(n.value - day - 1)} + 4px)`;
	return side === 'before' ? { right, top: '2px' } : { right, bottom: '2px' };
}

function hhmm(minute: number): string {
	const h = Math.floor(minute / 60);
	const m = minute % 60;
	return String(h).padStart(2, '0') + ':' + String(m).padStart(2, '0');
}

function dayTitle(day: number): string {
	return days.value[day]?.title ?? '';
}

function eventLabel(b: Interval & { title: string }): string {
	return `${b.title}, ${dayTitle(b.day)}, ${hhmm(b.from)} ${t('weekcal.to')} ${hhmm(b.to)}`;
}

function intervalText(iv: Interval): string {
	return `${dayTitle(iv.day)}, ${hhmm(iv.from)} ${t('weekcal.to')} ${hhmm(iv.to)}`;
}

// --- actions ---

// seq numbers the actions of the page, never twice, also over a new instance of the component
let seq = 0;

interface Pending {
	seq: number;
	preview: Interval | null;
}

const pending = ref<Pending | null>(null);

function busy(): boolean {
	return !!props.ui.disabled || !!pending.value;
}

function send(
	kind: 'click' | 'select' | 'move' | 'resize',
	id: string,
	iv: Interval,
	base: Interval | null,
	preview: Interval | null
) {
	if (!props.ui.input) return;
	seq = Math.max(seq, Number(props.ui.ack ?? 0)) + 1;
	pending.value = { seq, preview };
	const action: Record<string, unknown> = { seq, kind, id, day: iv.day, from: iv.from, to: iv.to };
	if (base) action.base = { day: base.day, from: base.from, to: base.to };
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.input, 0, nextRID(), JSON.stringify(action)));
}

function click(b: Block) {
	if (!canClick.value) return;
	const iv = { day: b.day, from: b.from, to: b.to };
	send('click', b.id, iv, iv, null);
}

function select(iv: Interval) {
	if (!canSelect.value) return;
	send('select', '', iv, null, iv);
}

// commit sends the change of a block, or a click, if it ends without a change
function commit(b: Block, kind: 'move' | 'resize', iv: Interval) {
	if (iv.day === b.day && iv.from === b.from && iv.to === b.to) {
		click(b);
		return;
	}
	send(kind, b.id, iv, { day: b.day, from: b.from, to: b.to }, iv);
}

// the tree with the number of the pending action or a higher one shows the state of the server
watch(
	() => Number(props.ui.ack ?? 0),
	(ack) => {
		if (pending.value && ack >= pending.value.seq) pending.value = null;
	}
);

// after a reconnect, the server shows its state: an action, which was on its way, is lost
let disconnected = false;

function onConnectionChanged(state: ConnectionState) {
	if (!state.connected) {
		disconnected = true;
		return;
	}
	if (disconnected) {
		disconnected = false;
		pending.value = null;
		cancelGesture();
		kbMove.value = null;
		selection.value = null;
	}
}

// --- previews ---

const announcement = ref('');

function announce(iv: Interval | null) {
	if (iv) announcement.value = intervalText(iv);
}

const shownPreview = computed<Interval | null>(() => {
	if (gesture.value?.preview) return gesture.value.preview;
	if (pending.value?.preview) return pending.value.preview;
	if (kbMove.value) return kbMove.value.iv;
	const s = selectionInterval.value;
	if (s) return s;
	return null;
});

function blockById(id: string): Block | undefined {
	return blocks.value.find((b) => b.id === id);
}

function blockAt(day: number, minute: number): Block | undefined {
	return blocks.value.find((b) => b.day === day && b.from <= minute && minute < b.to);
}

function movable(b: Block): boolean {
	return canMove.value && b.movable && !props.ui.disabled;
}

function floorStep(m: number): number {
	return Math.floor(m / step.value) * step.value;
}

function ceilStep(m: number): number {
	return Math.ceil(m / step.value) * step.value;
}

// spanned is the interval from the raster point before the beginning to the one after the pointer, at least one
// raster, within the range
function spanned(day: number, a: number, b: number): Interval {
	const lo = Math.min(a, b);
	const hi = Math.max(a, b);
	let from = Math.max(floorStep(lo), range.value.from);
	let to = Math.min(ceilStep(hi), range.value.to);
	if (to <= from) {
		if (from + step.value <= range.value.to) to = from + step.value;
		else from = to - step.value;
	}
	return { day, from, to };
}

// clickSpanAt is the interval of a click into free time: ClickSpan from the raster point before, cut at the range
function clickSpanAt(day: number, minute: number): Interval {
	const from = Math.max(Math.min(floorStep(minute), range.value.to - step.value), range.value.from);
	return { day, from, to: Math.min(from + clickSpan.value, range.value.to) };
}

// moved is the block moved by a number of minutes, rounded to the step, and days, within the range
function moved(b: Block, minutes: number, dayDelta: number): Interval {
	let delta = Math.round(minutes / step.value) * step.value;
	const lo = Math.ceil((range.value.from - b.from) / step.value) * step.value;
	const hi = Math.floor((range.value.to - b.to) / step.value) * step.value;
	delta = Math.min(Math.max(delta, lo), hi);
	const day = Math.min(Math.max(b.day + dayDelta, 0), n.value - 1);
	return { day, from: b.from + delta, to: b.to + delta };
}

// resized is the block, whose end is set to the raster point next to to, at least the first one after its beginning
function resized(b: Block, to: number): Interval {
	const min = floorStep(b.from) + step.value;
	const end = Math.min(Math.max(Math.round(to / step.value) * step.value, min), range.value.to);
	return { day: b.day, from: b.from, to: end };
}

// --- pointer ---

interface Gesture {
	pointerId: number;
	touch: boolean;
	x: number;
	y: number;
	day: number;
	minute: number;
	block: Block | null;
	resize: boolean;
	// pending until it is a drag, held after a long press on a block, which cannot be dragged
	phase: 'pending' | 'held' | 'dragging';
	preview: Interval | null;
	timer: number | null;
}

const gesture = ref<Gesture | null>(null);
const draggingId = computed(() => {
	const g = gesture.value;
	return g && g.phase === 'dragging' && g.block ? g.block.id : '';
});

function pointAt(e: PointerEvent): { day: number; minute: number; colWidth: number } {
	const rect = body.value!.getBoundingClientRect();
	const colWidth = rect.width / n.value;
	const day = Math.min(Math.max(Math.floor((e.clientX - rect.left) / colWidth), 0), n.value - 1);
	const minute = range.value.from + ((e.clientY - rect.top) / hourHeight.value) * 60;
	return { day, minute, colWidth };
}

function onPointerDown(e: PointerEvent) {
	if (gesture.value || busy()) return;
	if (e.pointerType !== 'touch' && e.button !== 0) return;
	const target = e.target as HTMLElement;
	const el = target.closest('[data-weekcal-event]') as HTMLElement | null;
	const block = el ? (blockById(el.dataset.weekcalEvent ?? '') ?? null) : null;
	if (block) {
		if (!canClick.value && !movable(block)) return;
	} else if (!canSelect.value) {
		return;
	}
	const p = pointAt(e);
	const touch = e.pointerType === 'touch';
	gesture.value = {
		pointerId: e.pointerId,
		touch,
		x: e.clientX,
		y: e.clientY,
		day: p.day,
		minute: p.minute,
		block,
		resize: !!block && movable(block) && !!target.closest('.resize'),
		phase: 'pending',
		preview: null,
		timer: null,
	};
	// an enclosing drag and drop, e.g. a draggable tile, must not start from within the calendar
	document.addEventListener('dragstart', preventDragStart, true);
	window.addEventListener('keydown', onGestureKeydown, true);
	if (touch) {
		gesture.value.timer = window.setTimeout(onHold, TOUCH_HOLD_MS);
	} else {
		try {
			body.value?.setPointerCapture(e.pointerId);
		} catch {
			// the pointer may be gone already
		}
	}
}

// onHold begins the drag of a finger, which has been held still for half a second
function onHold() {
	const g = gesture.value;
	if (!g || g.phase !== 'pending') return;
	g.timer = null;
	if (g.block && !movable(g.block)) {
		g.phase = 'held';
		return;
	}
	beginDrag(g, g.x, g.y);
}

function beginDrag(g: Gesture, x: number, y: number) {
	g.phase = 'dragging';
	updateDrag(g, x, y);
}

function updateDrag(g: Gesture, x: number, y: number) {
	const rect = body.value!.getBoundingClientRect();
	const colWidth = rect.width / n.value;
	const minutes = ((y - g.y) / hourHeight.value) * 60;
	let iv: Interval;
	if (!g.block) {
		iv = spanned(g.day, g.minute, g.minute + minutes);
	} else if (g.resize) {
		iv = resized(g.block, g.block.to + minutes);
	} else {
		iv = moved(g.block, minutes, Math.round((x - g.x) / colWidth));
	}
	const prev = g.preview;
	if (!prev || prev.day !== iv.day || prev.from !== iv.from || prev.to !== iv.to) {
		g.preview = iv;
		announce(iv);
	}
}

function onPointerMove(e: PointerEvent) {
	const g = gesture.value;
	if (!g || e.pointerId !== g.pointerId) return;
	const dist = Math.hypot(e.clientX - g.x, e.clientY - g.y);
	if (g.phase === 'pending') {
		if (g.touch) {
			// a finger, which moves before the long press, scrolls the page
			if (dist > TOUCH_SLOP_PX) cancelGesture();
			return;
		}
		if (dist < MOUSE_DRAG_PX) return;
		if (g.block && !movable(g.block)) {
			g.phase = 'held';
			return;
		}
		beginDrag(g, e.clientX, e.clientY);
		return;
	}
	if (g.phase === 'dragging') updateDrag(g, e.clientX, e.clientY);
}

function onPointerUp(e: PointerEvent) {
	const g = gesture.value;
	if (!g || e.pointerId !== g.pointerId) return;
	endGesture();
	if (g.phase === 'dragging' && g.preview) {
		if (g.block) commit(g.block, g.resize ? 'resize' : 'move', g.preview);
		else select(g.preview);
		return;
	}
	if (g.block) {
		click(g.block);
		return;
	}
	if (g.phase === 'pending') select(clickSpanAt(g.day, g.minute));
}

function onPointerCancel(e: PointerEvent) {
	const g = gesture.value;
	if (g && e.pointerId === g.pointerId) cancelGesture();
}

// a finger, which drags, must not scroll the page: only a cancelable touchmove keeps the pointer events. While a
// finger is held for the long press, the browser holds back its moves within its own slop; a move beyond it, before
// the long press, has cancelled the gesture in onPointerMove, which comes first, so the page scrolls.
function onTouchMove(e: TouchEvent) {
	const g = gesture.value;
	if (g && g.touch && g.phase === 'dragging' && e.cancelable) e.preventDefault();
}

function onGestureKeydown(e: KeyboardEvent) {
	if (e.key !== 'Escape' || !gesture.value) return;
	e.preventDefault();
	e.stopImmediatePropagation();
	cancelGesture();
}

function preventDragStart(e: DragEvent) {
	e.preventDefault();
}

function endGesture() {
	const g = gesture.value;
	if (g?.timer) window.clearTimeout(g.timer);
	if (g && !g.touch) {
		try {
			body.value?.releasePointerCapture(g.pointerId);
		} catch {
			// released already
		}
	}
	gesture.value = null;
	document.removeEventListener('dragstart', preventDragStart, true);
	window.removeEventListener('keydown', onGestureKeydown, true);
}

function cancelGesture() {
	if (gesture.value) endGesture();
}

function onContextMenu(e: MouseEvent) {
	// a long press of a finger opens no menu within the calendar
	if (gesture.value?.touch || (e.target as HTMLElement).closest('.body')) e.preventDefault();
}

// --- keyboard ---

const cursor = ref({ day: 0, minute: range.value.from });
// selection spans from the cursor to head with shift and the arrow keys
const selection = ref<{ head: number } | null>(null);
const selectionInterval = computed<Interval | null>(() => {
	const s = selection.value;
	if (!s || s.head === cursor.value.minute) return null;
	return {
		day: cursor.value.day,
		from: Math.min(s.head, cursor.value.minute),
		to: Math.max(s.head, cursor.value.minute),
	};
});
// kbMove is the preview of a block, which the keyboard moves or resizes
const kbMove = ref<{ id: string; kind: 'move' | 'resize'; iv: Interval } | null>(null);

// the cursor stays on a raster point of the range and a day, also if the range or the days change
watch(
	() => [range.value.from, range.value.to, step.value, n.value],
	() => {
		const c = cursor.value;
		const minute = Math.min(Math.max(floorStep(c.minute), range.value.from), range.value.to - step.value);
		const day = Math.min(c.day, n.value - 1);
		if (minute !== c.minute || day !== c.day) cursor.value = { day, minute };
	}
);

const ARROWS = ['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight'];

function onKeydown(e: KeyboardEvent) {
	const target = e.target as HTMLElement;
	const arrow = ARROWS.includes(e.key);
	// the calendar always takes Alt with an arrow, which is back and forward in Windows and Linux
	if (arrow && e.altKey) e.preventDefault();

	if (target === cursorEl.value) {
		onCursorKeydown(e, arrow);
		return;
	}
	const id = target.dataset?.weekcalEvent;
	if (id !== undefined) onBlockKeydown(e, id, arrow);
}

function onCursorKeydown(e: KeyboardEvent, arrow: boolean) {
	const c = cursor.value;
	const r = range.value;
	if (arrow && e.altKey) return;
	if (arrow && e.shiftKey) {
		e.preventDefault();
		if (e.key !== 'ArrowUp' && e.key !== 'ArrowDown') return;
		if (!selection.value) {
			// spanning begins in free time only
			if (busy() || !canSelect.value || blockAt(c.day, c.minute)) return;
			selection.value = { head: c.minute };
		}
		const head = selection.value.head + (e.key === 'ArrowDown' ? step.value : -step.value);
		selection.value = { head: Math.min(Math.max(head, r.from), r.to) };
		announce(selectionInterval.value);
		return;
	}
	if (arrow) {
		e.preventDefault();
		selection.value = null;
		let { day, minute } = c;
		if (e.key === 'ArrowUp') minute = Math.max(minute - step.value, r.from);
		if (e.key === 'ArrowDown') minute = Math.min(minute + step.value, r.to - step.value);
		if (e.key === 'ArrowLeft') day = Math.max(day - 1, 0);
		if (e.key === 'ArrowRight') day = Math.min(day + 1, n.value - 1);
		cursor.value = { day, minute };
		return;
	}
	if (e.key === 'Enter') {
		e.preventDefault();
		if (busy()) return;
		const b = blockAt(c.day, c.minute);
		const s = selectionInterval.value;
		if (b && !s) {
			click(b);
			return;
		}
		selection.value = null;
		select(s ?? clickSpanAt(c.day, c.minute));
		return;
	}
	if (e.key === 'Escape' && selection.value) {
		e.preventDefault();
		e.stopPropagation();
		selection.value = null;
	}
}

function onBlockKeydown(e: KeyboardEvent, id: string, arrow: boolean) {
	const b = blockById(id);
	if (!b) return;
	const r = range.value;
	if (e.key === 'Enter') {
		e.preventDefault();
		if (busy()) return;
		const m = kbMove.value;
		kbMove.value = null;
		if (m && m.id === id) commit(b, m.kind, m.iv);
		else click(b);
		return;
	}
	if (e.key === 'Escape' && kbMove.value) {
		e.preventDefault();
		e.stopPropagation();
		kbMove.value = null;
		return;
	}
	if (!arrow || !e.altKey || busy() || !movable(b)) return;
	const kind = e.shiftKey ? 'resize' : 'move';
	let m = kbMove.value && kbMove.value.id === id ? kbMove.value : null;
	if (m && m.kind !== kind) return;
	let iv = m ? m.iv : { day: b.day, from: b.from, to: b.to };
	if (kind === 'move') {
		if (e.key === 'ArrowUp' && iv.from - step.value >= r.from)
			iv = { ...iv, from: iv.from - step.value, to: iv.to - step.value };
		if (e.key === 'ArrowDown' && iv.to + step.value <= r.to)
			iv = { ...iv, from: iv.from + step.value, to: iv.to + step.value };
		if (e.key === 'ArrowLeft' && iv.day > 0) iv = { ...iv, day: iv.day - 1 };
		if (e.key === 'ArrowRight' && iv.day < n.value - 1) iv = { ...iv, day: iv.day + 1 };
	} else if (e.key === 'ArrowDown') {
		iv = { ...iv, to: Math.min(floorStep(iv.to) + step.value, r.to) };
	} else if (e.key === 'ArrowUp') {
		iv = { ...iv, to: Math.max(ceilStep(iv.to) - step.value, floorStep(iv.from) + step.value) };
	}
	m = { id, kind, iv };
	kbMove.value = m;
	announce(iv);
}

// --- focus ---

// focused is the element, which has the focus within the calendar: the cursor or a block by its id. It keeps the
// focus, also if a new tree moves the block within the DOM.
let focused: string | null = null;
let hadFocus = false;

function onFocusIn(e: FocusEvent) {
	const target = e.target as HTMLElement;
	if (target === cursorEl.value) focused = 'cursor';
	else if (target.dataset?.weekcalEvent !== undefined) focused = 'event:' + target.dataset.weekcalEvent;
}

function onFocusOut(e: FocusEvent) {
	const target = e.target as HTMLElement;
	const next = e.relatedTarget as Node | null;
	if (target === cursorEl.value && next !== target) selection.value = null;
	if (kbMove.value && target.dataset?.weekcalEvent === kbMove.value.id && next !== target) kbMove.value = null;
	if (next && root.value?.contains(next)) return;
	// the focus left the calendar, unless the element has only been moved within the DOM (onUpdated)
	window.setTimeout(() => {
		if (!root.value?.contains(document.activeElement)) focused = null;
	}, 0);
}

function focusedElement(): HTMLElement | null {
	if (!focused || !root.value) return null;
	if (focused === 'cursor') return cursorEl.value ?? null;
	const id = focused.slice('event:'.length);
	for (const el of root.value.querySelectorAll<HTMLElement>('[data-weekcal-event]')) {
		if (el.dataset.weekcalEvent === id) return el;
	}
	return null;
}

onBeforeUpdate(() => {
	hadFocus = !!root.value && root.value.contains(document.activeElement);
});

onUpdated(() => {
	if (!focused || !root.value) return;
	if (root.value.contains(document.activeElement)) return;
	if (!hadFocus) return;
	focusedElement()?.focus({ preventScroll: true });
});

// another calendar in the same place starts anew
watch(
	() => props.ui.id,
	() => {
		cancelGesture();
		pending.value = null;
		kbMove.value = null;
		selection.value = null;
	}
);

onMounted(() => {
	root.value?.addEventListener('touchmove', onTouchMove, { passive: false });
	ConnectionHandler.addConnectionChangeListener(onConnectionChanged);
});

onBeforeUnmount(() => {
	cancelGesture();
	root.value?.removeEventListener('touchmove', onTouchMove);
	ConnectionHandler.removeConnectionChangeListener(onConnectionChanged);
});
</script>

<style scoped>
@reference '@/assets/css/tailwind.css';

/* at most as wide as its place, also in a container, which follows the width of its content (gridStyle) */
.weekcal {
	--weekcal-line: rgb(from var(--M5) r g b / 0.6);
	--weekcal-hour-line: rgb(from var(--M5) r g b / 0.45);
	position: relative;
	overflow: auto;
	overscroll-behavior-x: contain;
	width: 100%;
	min-width: 0;
	max-width: 100%;
	touch-action: manipulation;
	@apply rounded-md border border-M5 bg-M1 text-M8;
}

.sizer {
	display: block;
	max-width: 100%;
	height: 0;
	pointer-events: none;
}

.message {
	position: sticky;
	left: 0;
	@apply bg-CBEB px-3 py-2 text-sm text-CBET;
}

.grid {
	display: grid;
}

.corner {
	grid-column: 1;
	grid-row: 1;
	position: sticky;
	top: 0;
	left: 0;
	z-index: 4;
	@apply border-b border-M5 bg-M1;
}

.head {
	grid-row: 1;
	position: sticky;
	top: 0;
	z-index: 3;
	min-width: 0;
	padding: 6px 8px;
	border-left: 1px solid var(--weekcal-line);
	@apply border-b border-M5 bg-M1 text-sm;
}

.head .title {
	font-weight: 600;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.head .subtitle {
	display: -webkit-box;
	-webkit-line-clamp: 2;
	-webkit-box-orient: vertical;
	overflow: hidden;
	white-space: pre-line;
	@apply text-xs text-M7;
}

.head[aria-current='date'] {
	background-color: rgb(from var(--I0) r g b / 0.12);
	@apply text-I0;
}

.head[data-muted] {
	background-color: rgb(from var(--M5) r g b / 0.35);
	@apply text-M7;
}

.axis {
	grid-column: 1;
	grid-row: 2;
	position: sticky;
	left: 0;
	z-index: 2;
	@apply bg-M1 text-xs text-M7;
}

.hour {
	position: absolute;
	right: 6px;
	transform: translateY(-50%);
	font-variant-numeric: tabular-nums;
	line-height: 1;
}

.hour.first {
	transform: translateY(2px);
}

.hour.last {
	transform: translateY(calc(-100% - 2px));
}

.body {
	grid-row: 2;
	position: relative;
	overflow: hidden;
	user-select: none;
	-webkit-user-select: none;
	-webkit-touch-callout: none;
	background-image: linear-gradient(to bottom, var(--weekcal-hour-line) 1px, transparent 1px);
	background-repeat: repeat-y;
}

.weekcal.pending .body {
	cursor: progress;
}

.col {
	position: absolute;
	top: 0;
	bottom: 0;
	border-left: 1px solid var(--weekcal-line);
	pointer-events: none;
}

.col.today {
	background-color: rgb(from var(--I0) r g b / 0.05);
}

.col.muted {
	background-color: rgb(from var(--M5) r g b / 0.2);
}

.cursor {
	position: absolute;
	box-sizing: border-box;
	pointer-events: none;
	outline: none;
	border-radius: 4px;
}

.cursor:focus-visible,
.cursor:focus {
	box-shadow: inset 0 0 0 2px var(--I0);
	background-color: rgb(from var(--I0) r g b / 0.08);
}

.event {
	--weekcal-event: var(--I0);
	position: absolute;
	box-sizing: border-box;
	overflow: hidden;
	padding: 2px 6px;
	border-radius: 4px;
	border-left: 3px solid var(--weekcal-event);
	background-color: color-mix(in srgb, var(--weekcal-event) 22%, var(--M1));
	box-shadow: inset 0 0 0 1px var(--M1);
	font-size: 0.75rem;
	line-height: 1.2;
	outline: none;
	z-index: 1;
}

.event.clickable,
.event.movable {
	cursor: pointer;
}

.event:focus-visible,
.event:focus {
	box-shadow:
		inset 0 0 0 1px var(--M1),
		0 0 0 2px var(--I0);
	z-index: 2;
}

.event[data-clipped~='top'] {
	border-top: 2px dashed var(--weekcal-event);
	border-top-left-radius: 0;
	border-top-right-radius: 0;
}

.event[data-clipped~='bottom'] {
	border-bottom: 2px dashed var(--weekcal-event);
	border-bottom-left-radius: 0;
	border-bottom-right-radius: 0;
}

.event[data-fixed] {
	background-image: repeating-linear-gradient(135deg, transparent 0 6px, rgb(from var(--M5) r g b / 0.25) 6px 8px);
}

.event[data-dragging] {
	opacity: 0.4;
}

.event-title {
	font-weight: 600;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.event-time {
	font-variant-numeric: tabular-nums;
	@apply text-M7;
}

.event-text {
	overflow: hidden;
	white-space: pre-line;
}

.resize {
	position: absolute;
	left: 0;
	right: 0;
	bottom: 0;
	cursor: ns-resize;
}

.hidden-count {
	position: absolute;
	pointer-events: none;
	z-index: 2;
	padding: 2px 5px;
	border-radius: 9999px;
	font-size: 0.7rem;
	font-weight: 600;
	line-height: 1;
	background-color: var(--M1);
	box-shadow: 0 0 0 1px var(--I0);
}

.hidden-count::before {
	margin-right: 2px;
}

.hidden-count.before::before {
	content: '▲';
}

.hidden-count.after::before {
	content: '▼';
}

.hidden-count {
	@apply text-I0;
}

.preview {
	position: absolute;
	box-sizing: border-box;
	pointer-events: none;
	z-index: 3;
	border-radius: 4px;
	border: 2px dashed var(--I0);
	background-color: rgb(from var(--I0) r g b / 0.18);
	padding: 2px 6px;
	font-size: 0.75rem;
	font-weight: 600;
	@apply text-I0;
}

.preview-time {
	font-variant-numeric: tabular-nums;
}

.now {
	position: absolute;
	height: 0;
	border-top: 2px solid var(--SE0);
	pointer-events: none;
	z-index: 3;
}

.now::before {
	content: '';
	position: absolute;
	left: -4px;
	top: -5px;
	width: 8px;
	height: 8px;
	border-radius: 50%;
	background-color: var(--SE0);
}
</style>
