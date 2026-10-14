<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div
		v-if="!ui.invisible"
		:id="ui.id"
		class="sheet"
		:style="frameStyles"
		@focusin="focused = true"
		@focusout="onFocusOut"
	>
		<div
			ref="scroller"
			class="scroller"
			tabindex="0"
			role="grid"
			:aria-rowcount="ui.rowCount ?? 0"
			:aria-colcount="ui.colCount ?? 0"
			@scroll.passive="onScroll"
			@keydown="onKeydown"
		>
			<div class="grid" :style="`width:${cols.total}px;height:${rows.total}px`">
				<div v-for="layer in layers" :key="layer.name" class="layer" :class="layer.name" :style="layer.style">
					<div
						v-for="c in layer.cells"
						:key="c.key"
						class="cell"
						:class="c.classes"
						:style="c.style"
						:title="c.error || undefined"
						role="gridcell"
						@pointerdown="onCellPointerDown($event, c.row, c.col)"
						@dblclick="onCellDblClick(c.row, c.col)"
					>
						<span v-if="c.checkbox !== undefined" class="checkbox" :class="{ checked: c.checkbox }">
							<svg v-if="c.checkbox" viewBox="0 0 16 16"><path d="M3.5 8.5l3 3 6-7" /></svg>
						</span>
						<span v-else class="text">{{ c.text }}</span>
					</div>

					<template v-if="editing && editing.layer === layer.name">
						<select
							v-if="editing.options"
							ref="editor"
							class="editor"
							:style="editing.style"
							:value="editing.value"
							@change="commitEdit(($event.target as HTMLSelectElement).value, 'down')"
							@keydown.stop="onEditorKeydown"
							@blur="cancelEdit"
						>
							<option v-for="o in editing.options" :key="o" :value="o">{{ o }}</option>
						</select>
						<input
							v-else
							ref="editor"
							class="editor"
							:style="editing.style"
							:value="editing.value"
							:inputmode="editing.numeric ? 'decimal' : 'text'"
							spellcheck="false"
							@keydown.stop="onEditorKeydown"
							@blur="commitEdit(($event.target as HTMLInputElement).value, 'none')"
						/>
					</template>
				</div>
			</div>
		</div>

		<!-- the error of the last edit and, on a touch device, copying without a keyboard -->
		<div v-if="notice || (coarse && focused)" class="footer">
			<span v-if="notice" class="notice" role="alert">{{ notice }}</span>
			<span class="grow"></span>
			<template v-if="coarse && focused">
				<span class="copied">{{ copied ? $t('sheet.copied') : $t('sheet.copy') }}</span>
				<button
					type="button"
					:aria-label="$t('sheet.copyCell')"
					@pointerdown.prevent
					@click="copyActive('cell')"
				>
					{{ $t('sheet.cell') }}
				</button>
				<button type="button" :aria-label="$t('sheet.copyRow')" @pointerdown.prevent @click="copyActive('row')">
					{{ $t('sheet.row') }}
				</button>
				<button type="button" :aria-label="$t('sheet.copyCol')" @pointerdown.prevent @click="copyActive('col')">
					{{ $t('sheet.col') }}
				</button>
			</template>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { Axis } from '@/components/sheet/axis';
import { type CopyCell, decimalSeparator, toHTML, toTSV } from '@/components/sheet/clipboard';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import {
	Sheet,
	SheetAlignValues,
	SheetBorder,
	SheetCell,
	SheetEditorValues,
	SheetLineStyleValues,
	SheetStyle,
	SheetValueTypeValues,
	SheetWrapValues,
	UpdateStateValueRequested,
} from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Sheet;
}>();

const serviceAdapter = useServiceAdapter();
const scroller = ref<HTMLElement>();
const editor = ref<HTMLInputElement[] | HTMLSelectElement[]>();

// --- geometry ---

const coarse = window.matchMedia?.('(pointer: coarse)').matches ?? false;
// a finger needs higher rows than a mouse pointer
const rowHeight = computed(() => Math.max(props.ui.rowHeight || 24, coarse ? 40 : 0));
const fontSize = computed(() => (rowHeight.value <= 24 ? 13 : 14));

const rows = computed(
	() => new Axis(props.ui.rowCount ?? 0, (props.ui.hiddenRows?.value ?? []).map(Number), rowHeight.value)
);
const cols = computed(() => {
	const sizes = new Map<number, number>();
	for (const s of props.ui.colWidths?.value ?? []) sizes.set(Number(s.index ?? 0), s.size ?? 0);
	return new Axis(
		props.ui.colCount ?? 0,
		(props.ui.hiddenCols?.value ?? []).map(Number),
		props.ui.colWidth || 100,
		sizes
	);
});

const viewWidth = ref(0);
const viewHeight = ref(0);
const scrollTop = ref(0);
const scrollLeft = ref(0);

const frozenRows = computed(() => Math.min(Number(props.ui.frozenRows ?? 0), rows.value.count));
// frozen columns, which take too much of a narrow screen, scroll with the others
const frozenCols = computed(() => {
	const n = Math.min(Number(props.ui.frozenCols ?? 0), cols.value.count);
	if (n > 0 && viewWidth.value > 0 && cols.value.span(0, n) > viewWidth.value * 0.4) return 0;
	return n;
});
const frozenHeight = computed(() => rows.value.span(0, frozenRows.value));
const frozenWidth = computed(() => cols.value.span(0, frozenCols.value));

const frameStyles = computed(() => frameCSS(props.ui.frame).join(';'));

// visible returns the range of the non-frozen rows and columns, which are visible, plus a small overscan.
const visible = computed(() => {
	const r = rows.value;
	const c = cols.value;
	const top = scrollTop.value + frozenHeight.value;
	const bottom = scrollTop.value + viewHeight.value;
	const left = scrollLeft.value + frozenWidth.value;
	const right = scrollLeft.value + viewWidth.value;
	return {
		rowFrom: Math.max(frozenRows.value, r.next(r.indexAtPosition(top), -2)),
		rowTo: r.next(r.indexAtPosition(bottom), 2) + 1,
		colFrom: Math.max(frozenCols.value, c.next(c.indexAtPosition(left), -1)),
		colTo: c.next(c.indexAtPosition(right), 1) + 1,
	};
});

// --- cells of the window ---

const cellMap = computed(() => {
	const m = new Map<string, SheetCell>();
	for (const c of props.ui.cells?.value ?? []) m.set(key(Number(c.row ?? 0), Number(c.col ?? 0)), c);
	return m;
});

function key(row: number, col: number): string {
	return row + ':' + col;
}

const merges = computed(() =>
	(props.ui.merges?.value ?? []).map((m) => ({
		row: Number(m.row ?? 0),
		col: Number(m.col ?? 0),
		rows: Number(m.rows ?? 1),
		cols: Number(m.cols ?? 1),
	}))
);

function mergeAt(row: number, col: number) {
	return merges.value.find((m) => row >= m.row && row < m.row + m.rows && col >= m.col && col < m.col + m.cols);
}

// optimistic values of edits, which the server has not processed yet
const optimistic = ref(new Map<string, { seq: number; value: string }>());
let editSeq = 0;

watch(
	() => Number(props.ui.editSeq ?? 0),
	(processed) => {
		for (const [k, v] of optimistic.value) {
			if (v.seq <= processed) optimistic.value.delete(k);
		}
	}
);

// --- selection ---

const anchor = ref({ row: 0, col: 0 });
const focus = ref({ row: 0, col: 0 });
const selecting = ref(false);
// the active cell is only marked, while the sheet has the focus
const focused = ref(false);

function onFocusOut(e: FocusEvent) {
	const sheet = e.currentTarget as HTMLElement;
	if (!sheet.contains(e.relatedTarget as Node | null)) focused.value = false;
}

const selection = computed(() => ({
	row: Math.min(anchor.value.row, focus.value.row),
	col: Math.min(anchor.value.col, focus.value.col),
	rowTo: Math.max(anchor.value.row, focus.value.row),
	colTo: Math.max(anchor.value.col, focus.value.col),
}));

// --- rendering ---

interface RenderedCell {
	key: string;
	row: number;
	col: number;
	text: string;
	checkbox?: boolean;
	error?: string;
	classes: Record<string, boolean>;
	style: string;
}

function lineCSS(b: SheetBorder | undefined, fallback: string): string {
	if (!b || !b.style || b.style === SheetLineStyleValues.SheetLineNone) return fallback;
	const style = ['none', 'solid', 'dashed', 'dotted', 'double'][b.style] ?? 'solid';
	const width = b.width || (b.style === SheetLineStyleValues.SheetLineDouble ? 3 : 1);
	return `${width}px ${style} ${b.color ? colorValue(b.color) : 'currentColor'}`;
}

const gridLine = computed(
	() => `1px solid ${props.ui.gridColor ? colorValue(props.ui.gridColor) : 'var(--sheet-grid)'}`
);

function renderCell(row: number, col: number): RenderedCell | null {
	const merge = mergeAt(row, col);
	if (merge && (merge.row !== row || merge.col !== col)) return null;

	const k = key(row, col);
	const cell = cellMap.value.get(k);
	const style: SheetStyle | undefined = cell?.style ? props.ui.styles?.value[Number(cell.style) - 1] : undefined;
	const width = merge ? cols.value.span(col, merge.cols) : cols.value.size(col);
	const height = merge ? rows.value.span(row, merge.rows) : rows.value.size(row);
	if (width <= 0 || height <= 0) return null;

	const pending = optimistic.value.get(k);
	let text = pending ? pending.value : (cell?.text ?? '');
	let checkbox: boolean | undefined;
	if (cell?.editor === SheetEditorValues.SheetEditorCheckbox) {
		checkbox = (pending ? pending.value : cell.value || cell.text) === 'true';
		text = '';
	}

	const numeric = cell?.type === SheetValueTypeValues.SheetValueNumber;
	let align = numeric ? 'end' : 'start';
	switch (style?.align) {
		case SheetAlignValues.SheetAlignStart:
			align = 'start';
			break;
		case SheetAlignValues.SheetAlignCenter:
			align = 'center';
			break;
		case SheetAlignValues.SheetAlignEnd:
			align = 'end';
			break;
	}

	const css = [
		`left:${cols.value.start(col)}px`,
		`top:${rows.value.start(row)}px`,
		`width:${width}px`,
		`height:${height}px`,
		`justify-content:${checkbox !== undefined ? 'center' : align}`,
		`text-align:${align}`,
		`border-right:${lineCSS(style?.borderRight, gridLine.value)}`,
		`border-bottom:${lineCSS(style?.borderBottom, gridLine.value)}`,
	];
	if (style?.borderTop?.style) css.push(`border-top:${lineCSS(style.borderTop, 'none')}`);
	if (style?.borderLeft?.style) css.push(`border-left:${lineCSS(style.borderLeft, 'none')}`);
	if (style?.backgroundColor) css.push(`background-color:${colorValue(style.backgroundColor)}`);
	if (style?.color) css.push(`color:${colorValue(style.color)}`);
	if (style?.bold) css.push('font-weight:600');
	if (style?.italic) css.push('font-style:italic');
	if (style?.underline) css.push('text-decoration:underline');
	if (style?.paddingX) css.push(`padding-left:${style.paddingX}px;padding-right:${style.paddingX}px`);
	css.push(`font-size:${style?.fontSize || fontSize.value}px`);

	const sel = selection.value;
	const lastRow = merge ? row + merge.rows - 1 : row;
	const lastCol = merge ? col + merge.cols - 1 : col;
	const selected = lastRow >= sel.row && row <= sel.rowTo && lastCol >= sel.col && col <= sel.colTo;

	return {
		key: k,
		row,
		col,
		text,
		checkbox,
		error: cell?.error,
		classes: {
			loading: !cell,
			selected: selected && (sel.row !== sel.rowTo || sel.col !== sel.colTo),
			active: focused.value && focusCell.value.row === row && focusCell.value.col === col,
			editable: !!cell?.editor && !props.ui.disabled,
			error: !!cell?.error,
			pending: !!pending,
			wrap: style?.wrap === SheetWrapValues.SheetWrapWrap,
			ellipsis: style?.wrap === SheetWrapValues.SheetWrapEllipsis,
		},
		style: css.join(';'),
	};
}

// the cell which is active, i.e. edited by keyboard, is the anchor of a merge, if it lies within one
const focusCell = computed(() => {
	const m = mergeAt(anchor.value.row, anchor.value.col);
	return m ? { row: m.row, col: m.col } : anchor.value;
});

function cellsOf(rowFrom: number, rowTo: number, colFrom: number, colTo: number): RenderedCell[] {
	const res: RenderedCell[] = [];
	const seen = new Set<string>();
	const push = (row: number, col: number) => {
		const k = key(row, col);
		if (seen.has(k)) return;
		seen.add(k);
		const c = renderCell(row, col);
		if (c) res.push(c);
	};

	for (let row = rowFrom; row < rowTo; row++) {
		if (rows.value.isHidden(row)) continue;
		for (let col = colFrom; col < colTo; col++) {
			if (cols.value.isHidden(col)) continue;
			const m = mergeAt(row, col);
			if (m) push(m.row, m.col);
			else push(row, col);
		}
	}

	return res;
}

const layers = computed(() => {
	const v = visible.value;
	const fr = frozenRows.value;
	const fc = frozenCols.value;
	const w = cols.value.total;
	const h = rows.value.total;
	return [
		{ name: 'body', style: `width:${w}px;height:${h}px`, cells: cellsOf(v.rowFrom, v.rowTo, v.colFrom, v.colTo) },
		{
			name: 'top',
			style: `width:${w}px;height:${frozenHeight.value}px`,
			cells: cellsOf(0, fr, v.colFrom, v.colTo),
		},
		{
			name: 'left',
			style: `width:${frozenWidth.value}px;height:${h}px`,
			cells: cellsOf(v.rowFrom, v.rowTo, 0, fc),
		},
		{
			name: 'corner',
			style: `width:${frozenWidth.value}px;height:${frozenHeight.value}px`,
			cells: cellsOf(0, fr, 0, fc),
		},
	];
});

function layerOf(row: number, col: number): string {
	const fr = row < frozenRows.value;
	const fc = col < frozenCols.value;
	if (fr && fc) return 'corner';
	if (fr) return 'top';
	if (fc) return 'left';
	return 'body';
}

// --- viewport ---

let lastRequested = '';
let frame: number | undefined;

function onScroll() {
	if (frame !== undefined) return;
	frame = requestAnimationFrame(() => {
		frame = undefined;
		measure();
	});
}

function measure() {
	const el = scroller.value;
	if (!el) return;
	scrollTop.value = el.scrollTop;
	scrollLeft.value = el.scrollLeft;
	viewWidth.value = el.clientWidth;
	viewHeight.value = el.clientHeight;
	requestWindow();
}

// requestWindow asks the server for a window around the visible area, once the visible area gets close to the
// edge of the window, which has been rendered.
function requestWindow() {
	if (!props.ui.viewport) return;
	const v = visible.value;
	const loaded = {
		row: Number(props.ui.windowRow ?? 0),
		col: Number(props.ui.windowCol ?? 0),
		rows: Number(props.ui.windowRows ?? 0),
		cols: Number(props.ui.windowCols ?? 0),
	};

	const pageRows = Math.max(10, v.rowTo - v.rowFrom);
	const pageCols = Math.max(4, v.colTo - v.colFrom);
	const marginRows = Math.floor(pageRows / 3);
	const marginCols = Math.floor(pageCols / 3);

	// rows and columns close to the visible area, which have not been rendered, except the frozen ones
	const missing =
		Math.max(v.rowFrom - marginRows, frozenRows.value) < loaded.row ||
		Math.min(v.rowTo + marginRows, rows.value.count) > loaded.row + loaded.rows ||
		Math.max(v.colFrom - marginCols, frozenCols.value) < loaded.col ||
		Math.min(v.colTo + marginCols, cols.value.count) > loaded.col + loaded.cols;
	if (!missing) return;

	const row = Math.max(0, v.rowFrom - pageRows);
	const col = Math.max(0, v.colFrom - pageCols);
	const request = `${row},${col},${v.rowTo - v.rowFrom + 2 * pageRows},${v.colTo - v.colFrom + 2 * pageCols}`;
	if (request === lastRequested) return;
	lastRequested = request;
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.viewport, 0, nextRID(), request));
}

// a new window may still not cover the visible area, e.g. after a long jump
watch(
	() => [props.ui.windowRow, props.ui.windowRows, props.ui.windowCol, props.ui.windowCols],
	() => {
		lastRequested = '';
		requestWindow();
	}
);

let resizeObserver: ResizeObserver | undefined;
onMounted(() => {
	measure();
	if (scroller.value && typeof ResizeObserver !== 'undefined') {
		resizeObserver = new ResizeObserver(() => measure());
		resizeObserver.observe(scroller.value);
	}
	document.addEventListener('copy', onCopy);
	document.addEventListener('pointerup', onPointerUp);
});

onUnmounted(() => {
	resizeObserver?.disconnect();
	document.removeEventListener('copy', onCopy);
	document.removeEventListener('pointerup', onPointerUp);
	document.removeEventListener('pointermove', onSelectMove);
	if (frame !== undefined) cancelAnimationFrame(frame);
	if (autoScroll !== undefined) cancelAnimationFrame(autoScroll);
});

// --- pointer ---

let lastTapped = '';

function onCellPointerDown(e: PointerEvent, row: number, col: number) {
	if (e.button !== 0) return;
	if (editing.value && !(editing.value.row === row && editing.value.col === col)) commitEditor();
	scroller.value?.focus({ preventScroll: true });

	const k = key(row, col);
	if (e.pointerType === 'touch') {
		// a second tap on the active cell edits it, a finger cannot select a range by dragging
		if (lastTapped === k && focusCell.value.row === row && focusCell.value.col === col) {
			startEdit(row, col);
		}
		lastTapped = k;
		anchor.value = { row, col };
		focus.value = { row, col };
		return;
	}

	const cell = cellMap.value.get(k);
	if (cell?.editor === SheetEditorValues.SheetEditorCheckbox && !e.shiftKey && !props.ui.disabled) {
		anchor.value = { row, col };
		focus.value = { row, col };
		toggle(row, col);
		return;
	}

	if (e.shiftKey) {
		focus.value = { row, col };
	} else {
		anchor.value = { row, col };
		focus.value = { row, col };
	}
	selecting.value = true;
	pointer = { x: e.clientX, y: e.clientY };
	document.addEventListener('pointermove', onSelectMove);
}

// a checkbox toggles on the first click already
function onCellDblClick(row: number, col: number) {
	if (!isCheckbox(row, col)) startEdit(row, col);
}

// --- selecting by dragging ---

let pointer: { x: number; y: number } | undefined;
let autoScroll: number | undefined;

// cellAt returns the cell under the point, or the nearest one, if the point lies outside of the sheet
function cellAt(x: number, y: number) {
	const el = scroller.value!;
	const rect = el.getBoundingClientRect();
	const lx = x - rect.left;
	const ly = y - rect.top;
	const posX =
		lx < frozenWidth.value
			? Math.max(0, lx)
			: el.scrollLeft + Math.min(Math.max(lx, frozenWidth.value), el.clientWidth - 1);
	const posY =
		ly < frozenHeight.value
			? Math.max(0, ly)
			: el.scrollTop + Math.min(Math.max(ly, frozenHeight.value), el.clientHeight - 1);
	return { row: rows.value.indexAtPosition(posY), col: cols.value.indexAtPosition(posX) };
}

function onSelectMove(e: PointerEvent) {
	pointer = { x: e.clientX, y: e.clientY };
	if (scroller.value) focus.value = cellAt(pointer.x, pointer.y);
	if (autoScroll === undefined) autoScroll = requestAnimationFrame(scrollWhileSelecting);
}

// scrollWhileSelecting scrolls, while the pointer is close to or beyond an edge, faster the further it is
function scrollWhileSelecting() {
	autoScroll = undefined;
	const el = scroller.value;
	if (!selecting.value || !pointer || !el) return;

	const rect = el.getBoundingClientRect();
	const edge = 24;
	const speed = (overshoot: number) => Math.min(40, Math.max(4, overshoot / 2));
	let dx = 0;
	let dy = 0;
	if (pointer.x < rect.left + frozenWidth.value + edge) dx = -speed(rect.left + frozenWidth.value + edge - pointer.x);
	else if (pointer.x > rect.left + el.clientWidth - edge) dx = speed(pointer.x - (rect.left + el.clientWidth - edge));
	if (pointer.y < rect.top + frozenHeight.value + edge) dy = -speed(rect.top + frozenHeight.value + edge - pointer.y);
	else if (pointer.y > rect.top + el.clientHeight - edge) dy = speed(pointer.y - (rect.top + el.clientHeight - edge));

	const before = { left: el.scrollLeft, top: el.scrollTop };
	el.scrollLeft += dx;
	el.scrollTop += dy;
	if (el.scrollLeft !== before.left || el.scrollTop !== before.top) {
		focus.value = cellAt(pointer.x, pointer.y);
		measure();
	}

	if (dx || dy) autoScroll = requestAnimationFrame(scrollWhileSelecting);
}

function onPointerUp() {
	selecting.value = false;
	pointer = undefined;
	document.removeEventListener('pointermove', onSelectMove);
	if (autoScroll !== undefined) cancelAnimationFrame(autoScroll);
	autoScroll = undefined;
}

// --- keyboard ---

function move(target: { row: number; col: number }, dRow: number, dCol: number) {
	const m = mergeAt(target.row, target.col);
	let row = target.row;
	let col = target.col;
	// leave a merge at its far side
	if (m && dRow > 0) row = m.row + m.rows - 1;
	if (m && dCol > 0) col = m.col + m.cols - 1;
	if (m && dRow < 0) row = m.row;
	if (m && dCol < 0) col = m.col;
	return {
		row: dRow ? rows.value.next(row, dRow) : target.row,
		col: dCol ? cols.value.next(col, dCol) : target.col,
	};
}

function onKeydown(e: KeyboardEvent) {
	if (editing.value) return;
	const ctrl = e.ctrlKey || e.metaKey;
	const page = Math.max(1, Math.floor((viewHeight.value - frozenHeight.value) / rowHeight.value) - 1);
	let delta: [number, number] | undefined;
	switch (e.key) {
		case 'ArrowUp':
			delta = [ctrl ? -rows.value.count : -1, 0];
			break;
		case 'ArrowDown':
			delta = [ctrl ? rows.value.count : 1, 0];
			break;
		case 'ArrowLeft':
			delta = [0, ctrl ? -cols.value.count : -1];
			break;
		case 'ArrowRight':
			delta = [0, ctrl ? cols.value.count : 1];
			break;
		case 'PageUp':
			delta = [-page, 0];
			break;
		case 'PageDown':
			delta = [page, 0];
			break;
		case 'Home':
			delta = [ctrl ? -rows.value.count : 0, -cols.value.count];
			break;
		case 'End':
			delta = [ctrl ? rows.value.count : 0, cols.value.count];
			break;
		case 'Tab':
			delta = [0, e.shiftKey ? -1 : 1];
			break;
		case 'Enter':
			if (isEditable(focusCell.value.row, focusCell.value.col)) {
				e.preventDefault();
				startEdit(focusCell.value.row, focusCell.value.col);
				return;
			}
			delta = [e.shiftKey ? -1 : 1, 0];
			break;
		case 'F2':
			e.preventDefault();
			startEdit(focusCell.value.row, focusCell.value.col);
			return;
		case 'Escape':
			focus.value = { ...anchor.value };
			return;
		case ' ':
			if (isCheckbox(focusCell.value.row, focusCell.value.col)) {
				e.preventDefault();
				toggle(focusCell.value.row, focusCell.value.col);
			}
			return;
		case 'Delete':
		case 'Backspace':
			if (
				isEditable(focusCell.value.row, focusCell.value.col) &&
				!isCheckbox(focusCell.value.row, focusCell.value.col)
			) {
				e.preventDefault();
				send(focusCell.value.row, focusCell.value.col, '');
			}
			return;
		case 'a':
			if (ctrl) {
				e.preventDefault();
				anchor.value = { row: rows.value.first(), col: cols.value.first() };
				focus.value = { row: rows.value.last(), col: cols.value.last() };
			}
			break;
	}

	if (delta) {
		e.preventDefault();
		const extend = e.shiftKey && e.key !== 'Tab' && e.key !== 'Enter';
		const from = extend ? focus.value : focusCell.value;
		const to = move(from, delta[0], delta[1]);
		if (extend) {
			focus.value = to;
		} else {
			anchor.value = to;
			focus.value = to;
		}
		scrollIntoView(to.row, to.col);
		return;
	}

	// typing starts editing and replaces the content, like in a spreadsheet application
	if (e.key.length === 1 && !ctrl && !e.altKey && isEditable(focusCell.value.row, focusCell.value.col)) {
		const cell = cellMap.value.get(key(focusCell.value.row, focusCell.value.col));
		if (
			cell?.editor === SheetEditorValues.SheetEditorText ||
			cell?.editor === SheetEditorValues.SheetEditorNumber
		) {
			e.preventDefault();
			startEdit(focusCell.value.row, focusCell.value.col, e.key);
		}
	}
}

function scrollIntoView(row: number, col: number) {
	const el = scroller.value;
	if (!el) return;
	if (row >= frozenRows.value) {
		const top = rows.value.start(row);
		const bottom = top + rows.value.size(row);
		if (top < el.scrollTop + frozenHeight.value) el.scrollTop = top - frozenHeight.value;
		else if (bottom > el.scrollTop + el.clientHeight) el.scrollTop = bottom - el.clientHeight;
	}
	if (col >= frozenCols.value) {
		const left = cols.value.start(col);
		const right = left + cols.value.size(col);
		if (left < el.scrollLeft + frozenWidth.value) el.scrollLeft = left - frozenWidth.value;
		else if (right > el.scrollLeft + el.clientWidth) el.scrollLeft = right - el.clientWidth;
	}
	measure();
}

// --- editing ---

interface Editing {
	row: number;
	col: number;
	layer: string;
	value: string;
	original: string;
	options?: string[];
	numeric: boolean;
	style: string;
}

const editing = ref<Editing>();

function isEditable(row: number, col: number): boolean {
	const cell = cellMap.value.get(key(row, col));
	return !props.ui.disabled && !!cell?.editor && cell.editor !== SheetEditorValues.SheetEditorNone;
}

function isCheckbox(row: number, col: number): boolean {
	return cellMap.value.get(key(row, col))?.editor === SheetEditorValues.SheetEditorCheckbox && isEditable(row, col);
}

function startEdit(row: number, col: number, typed?: string) {
	if (!isEditable(row, col)) return;
	const m = mergeAt(row, col);
	if (m) {
		row = m.row;
		col = m.col;
	}
	const cell = cellMap.value.get(key(row, col))!;
	if (cell.editor === SheetEditorValues.SheetEditorCheckbox) {
		toggle(row, col);
		return;
	}

	scrollIntoView(row, col);
	const pending = optimistic.value.get(key(row, col));
	const original = pending ? pending.value : cell.value || cell.text || '';
	const width = m ? cols.value.span(col, m.cols) : cols.value.size(col);
	const height = m ? rows.value.span(row, m.rows) : rows.value.size(row);
	editing.value = {
		row,
		col,
		layer: layerOf(row, col),
		value: typed ?? original,
		original,
		options: cell.editor === SheetEditorValues.SheetEditorSelect ? (cell.options?.value ?? []) : undefined,
		numeric: cell.editor === SheetEditorValues.SheetEditorNumber,
		style: `left:${cols.value.start(col)}px;top:${rows.value.start(row)}px;width:${Math.max(width, 80)}px;height:${height}px;font-size:${fontSize.value}px`,
	};

	nextTick(() => {
		const el = editor.value?.[0];
		el?.focus({ preventScroll: true });
		if (el instanceof HTMLInputElement && typed === undefined) el.select();
	});
}

function onEditorKeydown(e: KeyboardEvent) {
	const el = e.target as HTMLInputElement | HTMLSelectElement;
	switch (e.key) {
		case 'Enter':
			e.preventDefault();
			commitEdit(el.value, e.shiftKey ? 'up' : 'down');
			break;
		case 'Tab':
			e.preventDefault();
			commitEdit(el.value, e.shiftKey ? 'left' : 'right');
			break;
		case 'Escape':
			e.preventDefault();
			cancelEdit();
			break;
	}
}

function commitEditor() {
	const el = editor.value?.[0];
	if (el) commitEdit(el.value, 'none');
}

function commitEdit(value: string, then: 'up' | 'down' | 'left' | 'right' | 'none') {
	const e = editing.value;
	if (!e) return;
	editing.value = undefined;
	if (value !== e.original) send(e.row, e.col, value);

	const delta: Record<string, [number, number]> = {
		up: [-1, 0],
		down: [1, 0],
		left: [0, -1],
		right: [0, 1],
		none: [0, 0],
	};
	const [dRow, dCol] = delta[then];
	if (dRow || dCol) {
		const to = move({ row: e.row, col: e.col }, dRow, dCol);
		anchor.value = to;
		focus.value = to;
		scrollIntoView(to.row, to.col);
	}
	scroller.value?.focus({ preventScroll: true });
}

function cancelEdit() {
	if (!editing.value) return;
	editing.value = undefined;
	scroller.value?.focus({ preventScroll: true });
}

function toggle(row: number, col: number) {
	const cell = cellMap.value.get(key(row, col));
	const pending = optimistic.value.get(key(row, col));
	const current = (pending ? pending.value : cell?.value || cell?.text) === 'true';
	send(row, col, current ? 'false' : 'true');
}

function send(row: number, col: number, value: string) {
	if (!props.ui.edit) return;
	editSeq = Math.max(editSeq, Number(props.ui.editSeq ?? 0)) + 1;
	optimistic.value.set(key(row, col), { seq: editSeq, value });
	lastEdit.value = { row, col };
	serviceAdapter.sendEvent(
		new UpdateStateValueRequested(props.ui.edit, 0, nextRID(), JSON.stringify({ seq: editSeq, row, col, value }))
	);
}

// --- copy ---

function resolveColor(c?: string): string {
	if (!c) return '';
	const probe = document.createElement('span');
	probe.style.color = colorValue(c);
	document.body.appendChild(probe);
	const resolved = getComputedStyle(probe).color;
	probe.remove();
	return resolved;
}

interface CellRange {
	row: number;
	col: number;
	rowTo: number;
	colTo: number;
}

// loadedCells returns the cells of the range, if all of them have been rendered by the server
function loadedCells(sel: CellRange): (CopyCell | null)[][] | undefined {
	// the indices behind the frozen ones must lie within the window
	const loaded = (from: number, to: number, frozen: number, start: number, count: number) => {
		const first = Math.max(from, frozen);
		return first > to || (first >= start && to < start + count);
	};

	if (
		!loaded(
			sel.row,
			sel.rowTo,
			Number(props.ui.frozenRows ?? 0),
			Number(props.ui.windowRow ?? 0),
			Number(props.ui.windowRows ?? 0)
		) ||
		!loaded(
			sel.col,
			sel.colTo,
			Number(props.ui.frozenCols ?? 0),
			Number(props.ui.windowCol ?? 0),
			Number(props.ui.windowCols ?? 0)
		)
	) {
		return undefined;
	}

	const out: (CopyCell | null)[][] = [];
	for (let row = sel.row; row <= sel.rowTo; row++) {
		if (rows.value.isHidden(row)) continue;
		const line: (CopyCell | null)[] = [];
		for (let col = sel.col; col <= sel.colTo; col++) {
			if (cols.value.isHidden(col)) continue;
			const m = mergeAt(row, col);
			if (m && (m.row !== row || m.col !== col)) {
				line.push(null);
				continue;
			}
			const cell = cellMap.value.get(key(row, col));
			const style = cell?.style ? props.ui.styles?.value[Number(cell.style) - 1] : undefined;
			line.push({
				cell,
				style,
				rowSpan: m ? Math.min(m.rows, sel.rowTo - row + 1) : 1,
				colSpan: m ? Math.min(m.cols, sel.colTo - col + 1) : 1,
			});
		}
		out.push(line);
	}

	return out;
}

// copyOnServer lets the server render a large range, which has not been loaded, into the clipboard
function copyOnServer(sel: CellRange) {
	if (!props.ui.copy) return;
	const request = `${sel.row},${sel.col},${sel.rowTo - sel.row + 1},${sel.colTo - sel.col + 1}`;
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.copy, 0, nextRID(), request));
}

function clipboardData(out: (CopyCell | null)[][]) {
	const locale = document.documentElement.lang || navigator.language;
	return { text: toTSV(out, decimalSeparator(locale)), html: toHTML(out, resolveColor) };
}

function onCopy(e: ClipboardEvent) {
	const el = scroller.value;
	if (!el || editing.value || !el.contains(document.activeElement)) return;
	e.preventDefault();

	const out = loadedCells(selection.value);
	if (!out) {
		copyOnServer(selection.value);
		return;
	}

	const data = clipboardData(out);
	e.clipboardData?.setData('text/plain', data.text);
	e.clipboardData?.setData('text/html', data.html);
}

const copied = ref(false);

// copyActive copies the active cell, its row or its column, e.g. on a touch device without a keyboard
async function copyActive(what: 'cell' | 'row' | 'col') {
	const c = focusCell.value;
	const sel: CellRange = { row: c.row, col: c.col, rowTo: c.row, colTo: c.col };
	if (what === 'row') {
		sel.col = 0;
		sel.colTo = cols.value.count - 1;
	} else if (what === 'col') {
		sel.row = 0;
		sel.rowTo = rows.value.count - 1;
	}

	anchor.value = { row: sel.row, col: sel.col };
	focus.value = { row: sel.rowTo, col: sel.colTo };

	const out = loadedCells(sel);
	if (!out) {
		copyOnServer(sel);
		return;
	}

	const data = clipboardData(out);
	try {
		await navigator.clipboard.write([
			new ClipboardItem({
				'text/plain': new Blob([data.text], { type: 'text/plain' }),
				'text/html': new Blob([data.html], { type: 'text/html' }),
			}),
		]);
	} catch {
		await navigator.clipboard?.writeText(data.text);
	}

	copied.value = true;
	setTimeout(() => (copied.value = false), 1500);
}

// --- notice ---

// lastEdit is the cell of the last edit, whose error stays visible, also after the active cell has moved on
const lastEdit = ref<{ row: number; col: number }>();

const notice = computed(() => {
	for (const c of [lastEdit.value, focused.value ? focusCell.value : undefined]) {
		if (!c) continue;
		const error = cellMap.value.get(key(c.row, c.col))?.error;
		if (error) return `${cellName(c.row, c.col)}: ${error}`;
	}
	return '';
});

// cellName names a cell like the user sees it: by the frozen first column and the frozen header row, if any
function cellName(row: number, col: number): string {
	const rowName = Number(props.ui.frozenCols ?? 0) > 0 && col > 0 ? cellMap.value.get(key(row, 0))?.text : '';
	const colName = Number(props.ui.frozenRows ?? 0) > 0 && row > 0 ? cellMap.value.get(key(0, col))?.text : '';
	return `${rowName || row + 1} · ${colName || col + 1}`;
}
</script>

<style scoped>
.sheet {
	--sheet-grid: rgb(from var(--M5) r g b / 0.5);
	@apply relative flex flex-col overflow-hidden rounded-md border border-M5;
}

.footer {
	@apply flex flex-wrap items-center gap-2 border-t border-M5 px-2 py-1 text-sm;

	.notice {
		@apply text-SE0;
	}

	.grow {
		@apply flex-1;
	}

	.copied {
		@apply text-ST0;
	}

	button {
		@apply rounded-md border border-M5 px-2 py-1;
	}
}

.scroller {
	@apply w-full min-h-0 flex-1 overflow-auto outline-none;
	overscroll-behavior: contain;
}

.grid {
	display: grid;
	position: relative;
}

.layer {
	grid-area: 1 / 1;
	position: relative;
}

.layer.top,
.layer.left,
.layer.corner {
	position: sticky;
	z-index: 2;
	pointer-events: none;
}

.layer.top {
	top: 0;
	align-self: start;
}

.layer.left {
	left: 0;
	justify-self: start;
}

.layer.corner {
	top: 0;
	left: 0;
	align-self: start;
	justify-self: start;
	z-index: 3;
}

.layer.top .cell,
.layer.left .cell,
.layer.corner .cell {
	pointer-events: auto;
}

/* frozen rows are usually a header */
.layer.top .cell,
.layer.corner .cell {
	background-color: var(--M2);
	font-weight: 600;
}

.cell {
	position: absolute;
	box-sizing: border-box;
	display: flex;
	align-items: center;
	padding: 0 6px;
	overflow: hidden;
	white-space: nowrap;
	background-color: var(--M1);
	user-select: none;
	cursor: cell;
}

.cell .text {
	overflow: hidden;
	min-width: 0;
}

.cell.ellipsis .text {
	text-overflow: ellipsis;
}

.cell.wrap {
	white-space: normal;
	align-items: flex-start;
	padding-top: 2px;
	line-height: 1.25;
}

.cell.wrap .text {
	overflow-wrap: anywhere;
}

.cell.loading::after {
	content: '';
	position: absolute;
	left: 6px;
	right: 6px;
	top: 40%;
	height: 20%;
	border-radius: 2px;
	background: rgb(from var(--M5) r g b / 0.15);
}

.cell.selected {
	box-shadow: inset 0 0 0 9999px rgb(from var(--I0) r g b / 0.12);
}

.cell.active {
	outline: 2px solid var(--I0);
	outline-offset: -2px;
	z-index: 1;
}

.cell.error {
	box-shadow: inset 0 0 0 2px var(--SE0);
}

.cell.pending {
	opacity: 0.6;
}

.checkbox {
	@apply size-4 rounded-sm border border-M7 flex items-center justify-center;

	&.checked {
		@apply bg-I0 border-I0;
	}

	svg {
		@apply size-3;
		fill: none;
		stroke: var(--M1);
		stroke-width: 2;
	}
}

.editor {
	position: absolute;
	z-index: 4;
	box-sizing: border-box;
	padding: 0 5px;
	border: 2px solid var(--I0);
	background: var(--M1);
	color: var(--M8);
	outline: none;
	pointer-events: auto;
}
</style>
