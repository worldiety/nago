<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div v-if="!ui.invisible" :id="ui.id" class="color-picker" :class="{ disabled: ui.disabled }" :style="frameStyles">
		<span v-if="ui.label" class="label" :class="{ 'text-SE0': ui.errorText }">{{ ui.label }}</span>

		<!-- preview, hex and mode -->
		<div class="top-row">
			<div class="checker preview">
				<div class="fill" :style="`background-color: ${hex}`"></div>
			</div>
			<input
				class="field hex"
				:value="hex"
				:disabled="ui.disabled"
				spellcheck="false"
				aria-label="Hex"
				@change="onHexChange"
			/>
			<button
				type="button"
				class="icon-button"
				:title="copied ? $t('colorPicker.copied') : $t('colorPicker.copy')"
				:aria-label="$t('colorPicker.copy')"
				@click="copy"
			>
				<svg v-if="!copied" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M15.75 17.25v3.375c0 .621-.504 1.125-1.125 1.125h-9.75a1.125 1.125 0 0 1-1.125-1.125V7.875c0-.621.504-1.125 1.125-1.125H6.75a9.06 9.06 0 0 1 1.5.124m7.5 10.376h3.375c.621 0 1.125-.504 1.125-1.125V11.25c0-4.46-3.243-8.161-7.5-9.376a9.06 9.06 0 0 0-1.5-.124H9.375c-.621 0-1.125.504-1.125 1.125v3.5m7.5 10.375H9.375a1.125 1.125 0 0 1-1.125-1.125v-9.25m12 6.625v-1.875a3.375 3.375 0 0 0-3.375-3.375h-1.5a1.125 1.125 0 0 1-1.125-1.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H9.75"
					/>
				</svg>
				<svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
				</svg>
			</button>
			<div class="modes" role="radiogroup" :aria-label="$t('colorPicker.mode')">
				<button
					v-for="m in modes"
					:key="m"
					type="button"
					role="radio"
					:aria-checked="mode === m"
					:class="{ active: mode === m }"
					:disabled="ui.disabled"
					@click="switchMode(m)"
				>
					{{ m.toUpperCase() }}
				</button>
			</div>
		</div>

		<!-- the handles may stand out by half of their size, so that the extremes can be picked -->
		<div class="sliders">
			<!-- the 2D area: chroma and tone for HCT, saturation and value for HSV -->
			<div
				class="area"
				:class="{ hatched: mode === 'hct' }"
				:style="areaStyles"
				role="slider"
				:aria-label="$t('colorPicker.area')"
				:aria-valuetext="areaValueText"
				:tabindex="ui.disabled ? -1 : 0"
				@pointerdown="startDrag($event, onArea)"
				@keydown="onAreaKey"
			>
				<canvas
					v-show="mode === 'hct'"
					ref="canvas"
					:width="canvasWidth"
					:height="canvasHeight"
					:style="mode === 'hct' ? gamutClip : ''"
				></canvas>
				<div class="handle" :style="areaHandleStyles"></div>
			</div>

			<!-- hue -->
			<div
				class="bar"
				:style="`background: ${hueGradient}`"
				role="slider"
				:aria-label="$t('colorPicker.hue')"
				aria-valuemin="0"
				aria-valuemax="360"
				:aria-valuenow="Math.round(hue)"
				:tabindex="ui.disabled ? -1 : 0"
				@pointerdown="startDrag($event, onHue)"
				@keydown="onBarKey($event, 'hue')"
			>
				<div class="handle" :style="`left: ${(hue / 360) * 100}%; background-color: ${hueColor}`"></div>
			</div>

			<!-- alpha -->
			<div
				v-if="!ui.disableAlpha"
				class="bar checker"
				role="slider"
				:aria-label="$t('colorPicker.alpha')"
				aria-valuemin="0"
				aria-valuemax="100"
				:aria-valuenow="Math.round(alpha * 100)"
				:tabindex="ui.disabled ? -1 : 0"
				@pointerdown="startDrag($event, onAlpha)"
				@keydown="onBarKey($event, 'alpha')"
			>
				<div class="fill" :style="`background: linear-gradient(to right, transparent, ${opaqueHex})`"></div>
				<div class="handle checker" :style="`left: ${alpha * 100}%`">
					<div class="fill" :style="`background-color: ${hex}`"></div>
				</div>
			</div>
		</div>

		<!-- notations -->
		<div class="notations">
			<label v-for="n in notations" :key="n.name" class="notation" :class="{ wide: n.name === 'CMYK' }">
				<span>{{ n.name }}</span>
				<input
					class="field"
					:value="n.text"
					:disabled="ui.disabled"
					spellcheck="false"
					@change="onNotationChange(n.name, $event)"
				/>
			</label>
		</div>

		<p v-if="ui.errorText" class="text-sm text-error">{{ ui.errorText }}</p>
		<p v-else-if="ui.supportingText" class="text-sm text-disabled-text">{{ ui.supportingText }}</p>
	</div>
</template>

<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import {
	type HCT,
	type HSV,
	type RGB,
	cmykToRgb,
	hctToRgb,
	hslToRgb,
	hsvToRgb,
	maxChroma,
	parseHex,
	parseNumbers,
	rgbToCmyk,
	rgbToHct,
	rgbToHsl,
	rgbToHsv,
	toHex,
	wrapHue,
} from '@/components/colorpicker/colorspaces';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { ColorPicker, ColorPickerModeValues, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

type Mode = 'hct' | 'hsv';

const props = defineProps<{
	ui: ColorPicker;
}>();

const serviceAdapter = useServiceAdapter();

const modes: Mode[] = ['hct', 'hsv'];
const mode = ref<Mode>(props.ui.mode === ColorPickerModeValues.ColorPickerModeHSV ? 'hsv' : 'hct');

// The color is kept in the coordinates of the active mode, otherwise the hue would jump for grays. For HCT, the
// chroma is the requested one, which may be more than the hue and tone can display, so that it comes back, if the
// user moves the hue back.
const hct = reactive<HCT>({ h: 0, c: 0, t: 50 });
const hsv = reactive<HSV>({ h: 0, s: 0, v: 50 });
const alpha = ref(1);
const dragging = ref(false);
const copied = ref(false);

const canvas = ref<HTMLCanvasElement>();
const canvasWidth = 120;
const canvasHeight = 80;

const alphaEnabled = computed(() => !props.ui.disableAlpha);

const rgb = computed<RGB>(() => (mode.value === 'hct' ? hctToRgb(hct) : hsvToRgb(hsv)));
const hex = computed(() => toHex(rgb.value, alphaEnabled.value ? alpha.value : 1));
const opaqueHex = computed(() => toHex(rgb.value));
const hue = computed(() => (mode.value === 'hct' ? hct.h : hsv.h));

const frameStyles = computed(() => frameCSS(props.ui.frame).join(';'));

// the scale of the chroma axis for the current hue, so that the area is used well for every hue
const chromaScaleCache = new Map<number, number>();
function chromaScale(h: number): number {
	const key = Math.round(h);
	let scale = chromaScaleCache.get(key);
	if (scale === undefined) {
		scale = 1;
		for (let t = 2; t < 100; t += 2) {
			scale = Math.max(scale, maxChroma(key, t));
		}
		scale = Math.ceil(scale);
		chromaScaleCache.set(key, scale);
	}

	return scale;
}

const displayedChroma = computed(() => Math.min(hct.c, maxChroma(hct.h, hct.t)));

const areaHandleStyles = computed(() => {
	let x: number, y: number;
	if (mode.value === 'hct') {
		x = displayedChroma.value / chromaScale(hct.h);
		y = 1 - hct.t / 100;
	} else {
		x = hsv.s / 100;
		y = 1 - hsv.v / 100;
	}

	return `left: ${x * 100}%; top: ${y * 100}%; background-color: ${opaqueHex.value}`;
});

const areaStyles = computed(() => {
	if (mode.value === 'hct') return '';
	return `background: linear-gradient(to top, #000, transparent), linear-gradient(to right, #fff, transparent), hsl(${hsv.h} 100% 50%)`;
});

const areaValueText = computed(() => {
	if (mode.value === 'hct') return `Chroma ${Math.round(displayedChroma.value)}, Tone ${Math.round(hct.t)}`;
	return `Saturation ${Math.round(hsv.s)}%, Value ${Math.round(hsv.v)}%`;
});

const hueGradient = computed(() => {
	const stops: string[] = [];
	for (let h = 0; h <= 360; h += 15) {
		stops.push(mode.value === 'hct' ? toHex(hctToRgb({ h, c: 200, t: 60 })) : `hsl(${h} 100% 50%)`);
	}

	return `linear-gradient(to right, ${stops.join(', ')})`;
});

const hueColor = computed(() =>
	mode.value === 'hct' ? toHex(hctToRgb({ h: hct.h, c: 200, t: 60 })) : `hsl(${hsv.h} 100% 50%)`
);

const round = (v: number) => Math.round(v);

const notations = computed(() => {
	const c = rgb.value;
	const hsl = rgbToHsl(c);
	const cmyk = rgbToCmyk(c);
	const h = mode.value === 'hct' ? { h: hct.h, c: displayedChroma.value, t: hct.t } : rgbToHct(c);
	const v = mode.value === 'hsv' ? hsv : rgbToHsv(c);
	const list = [
		{ name: 'HCT', text: `${round(h.h)}°, ${round(h.c)}, ${round(h.t)}` },
		{ name: 'RGB', text: `${round(c.r)}, ${round(c.g)}, ${round(c.b)}` },
		{ name: 'HSV', text: `${round(v.h)}°, ${round(v.s)}%, ${round(v.v)}%` },
		{ name: 'HSL', text: `${round(hsl.h)}°, ${round(hsl.s)}%, ${round(hsl.l)}%` },
		{ name: 'CMYK', text: `${round(cmyk.c)}%, ${round(cmyk.m)}%, ${round(cmyk.y)}%, ${round(cmyk.k)}%` },
	];
	if (alphaEnabled.value) list.push({ name: 'Alpha', text: `${round(alpha.value * 100)}%` });
	return list;
});

// setColor takes a color from outside of the active coordinates and keeps the hue, if the color has none.
function setColor(c: RGB, a?: number) {
	const h = rgbToHct(c);
	const keepHctHue = h.c < 1;
	Object.assign(hct, { h: keepHctHue ? hct.h : h.h, c: h.c, t: h.t });

	const v = rgbToHsv(c);
	const keepHsvHue = v.s === 0 || v.v === 0;
	Object.assign(hsv, { h: keepHsvHue ? hsv.h : v.h, s: v.s, v: v.v });

	if (a !== undefined) alpha.value = a;
}

function switchMode(m: Mode) {
	if (mode.value === m) return;
	const current = rgb.value;
	mode.value = m;
	setColor(current);
}

// --- server ---

let lastSent: string | undefined;

function commit() {
	if (!props.ui.inputValue || props.ui.disabled) return;
	const value = hex.value;
	if (value === lastSent) return;
	lastSent = value;
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), value));
}

let commitTimer: number | undefined;
function commitSoon() {
	clearTimeout(commitTimer);
	commitTimer = window.setTimeout(commit, 400);
}

function applyValue(value: string | undefined) {
	const parsed = parseHex(value);
	if (!parsed) return;
	const normalized = toHex(parsed.rgb, alphaEnabled.value ? parsed.alpha : 1);
	lastSent = normalized;
	if (normalized !== hex.value) setColor(parsed.rgb, parsed.alpha);
}

applyValue(props.ui.value);

watch(
	() => props.ui.value,
	(value) => {
		// the server must not move the handle under the pointer of the user
		if (!dragging.value) applyValue(value);
	}
);

// --- pointer and keyboard ---

const clamp01 = (v: number) => Math.min(1, Math.max(0, v));

function startDrag(e: PointerEvent, apply: (x: number, y: number) => void) {
	if (props.ui.disabled || e.button !== 0) return;
	const el = e.currentTarget as HTMLElement;
	e.preventDefault();
	// the focus enables the keyboard, but the ring is only for users of the keyboard
	const focusOptions = { preventScroll: true, focusVisible: false };
	el.focus(focusOptions);
	el.setPointerCapture(e.pointerId);
	dragging.value = true;

	const update = (ev: PointerEvent) => {
		const r = el.getBoundingClientRect();
		apply(clamp01((ev.clientX - r.left) / r.width), clamp01((ev.clientY - r.top) / r.height));
	};

	const end = () => {
		el.removeEventListener('pointermove', update);
		el.removeEventListener('pointerup', end);
		el.removeEventListener('pointercancel', end);
		dragging.value = false;
		commit();
	};

	update(e);
	el.addEventListener('pointermove', update);
	el.addEventListener('pointerup', end);
	el.addEventListener('pointercancel', end);
}

function onArea(x: number, y: number) {
	if (mode.value === 'hct') {
		hct.t = (1 - y) * 100;
		// the handle stays on the boundary of sRGB instead of picking a different color silently
		hct.c = Math.min(x * chromaScale(hct.h), maxChroma(hct.h, hct.t));
	} else {
		hsv.s = x * 100;
		hsv.v = (1 - y) * 100;
	}
}

function onHue(x: number) {
	const h = Math.min(x * 360, 359.9);
	if (mode.value === 'hct') hct.h = h;
	else hsv.h = h;
}

function onAlpha(x: number) {
	alpha.value = x;
}

function arrow(e: KeyboardEvent): [number, number] | undefined {
	const step = e.shiftKey ? 10 : 1;
	switch (e.key) {
		case 'ArrowLeft':
			return [-step, 0];
		case 'ArrowRight':
			return [step, 0];
		case 'ArrowUp':
			return [0, step];
		case 'ArrowDown':
			return [0, -step];
	}

	return undefined;
}

function onAreaKey(e: KeyboardEvent) {
	const d = arrow(e);
	if (!d || props.ui.disabled) return;
	e.preventDefault();
	if (mode.value === 'hct') {
		hct.t = Math.min(100, Math.max(0, hct.t + d[1]));
		const max = maxChroma(hct.h, hct.t);
		hct.c = Math.min(max, Math.max(0, Math.min(hct.c, max) + d[0]));
	} else {
		hsv.s = Math.min(100, Math.max(0, hsv.s + d[0]));
		hsv.v = Math.min(100, Math.max(0, hsv.v + d[1]));
	}

	commitSoon();
}

function onBarKey(e: KeyboardEvent, bar: 'hue' | 'alpha') {
	const d = arrow(e);
	if (!d || props.ui.disabled) return;
	e.preventDefault();
	const delta = d[0] + d[1];
	if (bar === 'hue') {
		const h = wrapHue(hue.value + delta);
		if (mode.value === 'hct') hct.h = h;
		else hsv.h = h;
	} else {
		alpha.value = clamp01(alpha.value + delta / 100);
	}

	commitSoon();
}

// --- fields ---

function onHexChange(e: Event) {
	const input = e.target as HTMLInputElement;
	const parsed = parseHex(input.value);
	if (parsed) {
		setColor(parsed.rgb, alphaEnabled.value ? parsed.alpha : 1);
		commit();
	}

	input.value = hex.value;
}

function onNotationChange(name: string, e: Event) {
	const input = e.target as HTMLInputElement;
	const text = input.value;
	switch (name) {
		case 'HCT': {
			const n = parseNumbers(text, 3);
			if (n) {
				if (mode.value === 'hct')
					Object.assign(hct, { h: wrapHue(n[0]), c: Math.max(0, n[1]), t: clamp(n[2], 0, 100) });
				else setColor(hctToRgb({ h: n[0], c: n[1], t: n[2] }));
			}
			break;
		}
		case 'RGB': {
			const n = parseNumbers(text, 3);
			if (n) setColor({ r: clamp(n[0], 0, 255), g: clamp(n[1], 0, 255), b: clamp(n[2], 0, 255) });
			break;
		}
		case 'HSV': {
			const n = parseNumbers(text, 3);
			if (n) {
				const v = { h: wrapHue(n[0]), s: clamp(n[1], 0, 100), v: clamp(n[2], 0, 100) };
				if (mode.value === 'hsv') Object.assign(hsv, v);
				else setColor(hsvToRgb(v));
			}
			break;
		}
		case 'HSL': {
			const n = parseNumbers(text, 3);
			if (n) setColor(hslToRgb({ h: wrapHue(n[0]), s: clamp(n[1], 0, 100), l: clamp(n[2], 0, 100) }));
			break;
		}
		case 'CMYK': {
			const n = parseNumbers(text, 4);
			if (n)
				setColor(
					cmykToRgb({
						c: clamp(n[0], 0, 100),
						m: clamp(n[1], 0, 100),
						y: clamp(n[2], 0, 100),
						k: clamp(n[3], 0, 100),
					})
				);
			break;
		}
		case 'Alpha': {
			const n = parseNumbers(text, 1);
			if (n) alpha.value = clamp(n[0], 0, 100) / 100;
			break;
		}
	}

	commit();
	// show the normalized notation, also if the input has been invalid
	nextTick(() => {
		input.value = notations.value.find((it) => it.name === name)?.text ?? text;
	});
}

function clamp(v: number, min: number, max: number): number {
	return Math.min(max, Math.max(min, v));
}

async function copy() {
	try {
		await navigator.clipboard.writeText(hex.value);
		copied.value = true;
		setTimeout(() => (copied.value = false), 1500);
	} catch {
		// the browser may deny the clipboard, the hex field can still be selected by hand
	}
}

// --- HCT area ---

let frame: number | undefined;

function scheduleDraw() {
	if (mode.value !== 'hct' || frame !== undefined) return;
	frame = requestAnimationFrame(() => {
		frame = undefined;
		draw();
	});
}

// draw renders chroma and tone of the current hue. The colors beyond sRGB are clamped to the boundary and cut off by
// gamutClip, which is sharp, although the canvas has only a low resolution.
function draw() {
	const ctx = canvas.value?.getContext('2d');
	if (!ctx) return;
	const h = hct.h;
	const scale = chromaScale(h);
	const img = ctx.createImageData(canvasWidth, canvasHeight);
	const data = img.data;
	for (let y = 0; y < canvasHeight; y++) {
		const t = (1 - (y + 0.5) / canvasHeight) * 100;
		const max = maxChroma(h, t);
		const boundary = hctToRgb({ h, c: max, t });
		for (let x = 0; x < canvasWidth; x++) {
			const c = ((x + 0.5) / canvasWidth) * scale;
			const p = c >= max ? boundary : hctToRgb({ h, c, t });
			const i = (y * canvasWidth + x) * 4;
			data[i] = p.r;
			data[i + 1] = p.g;
			data[i + 2] = p.b;
			data[i + 3] = 255;
		}
	}

	ctx.putImageData(img, 0, 0);
}

// gamutClip cuts the area to the colors, which the current hue can display in sRGB.
const gamutClip = computed(() => {
	const h = hct.h;
	const scale = chromaScale(h);
	const points = ['0% 0%'];
	for (let t = 100; t >= 0; t--) {
		points.push(`${((maxChroma(h, t) / scale) * 100).toFixed(2)}% ${100 - t}%`);
	}

	points.push('0% 100%');
	return `clip-path: polygon(${points.join(', ')})`;
});

watch([() => hct.h, mode], scheduleDraw);
onMounted(scheduleDraw);
onUnmounted(() => {
	if (frame !== undefined) cancelAnimationFrame(frame);
	clearTimeout(commitTimer);
});
</script>

<style scoped>
.color-picker {
	@apply flex flex-col gap-3 min-w-64;

	&.disabled {
		.area,
		.bar,
		.modes {
			@apply opacity-50 pointer-events-none;
		}
	}
}

.label {
	@apply text-base;
}

.checker {
	background-color: #fff;
	background-image:
		linear-gradient(45deg, #ccc 25%, transparent 25%), linear-gradient(-45deg, #ccc 25%, transparent 25%),
		linear-gradient(45deg, transparent 75%, #ccc 75%), linear-gradient(-45deg, transparent 75%, #ccc 75%);
	background-size: 12px 12px;
	background-position:
		0 0,
		0 6px,
		6px -6px,
		-6px 0;
}

.fill {
	@apply absolute inset-0 rounded-[inherit];
}

.sliders {
	@apply flex flex-col gap-3 px-2.5;
}

.top-row {
	@apply flex items-center gap-2;
}

.preview {
	@apply relative size-10 shrink-0 rounded-lg border border-M5 overflow-hidden;
}

.field {
	@apply min-w-0 rounded-md border border-M5 bg-transparent px-2 py-1.5 text-sm tabular-nums;

	&:focus {
		@apply outline-none border-I0;
	}
}

.hex {
	@apply flex-1 font-mono;
}

.icon-button {
	@apply size-9 shrink-0 p-2 rounded-md;

	&:hover,
	&:focus-visible {
		@apply bg-I0 bg-opacity-10;
	}
}

.modes {
	@apply flex shrink-0 rounded-md border border-M5 overflow-hidden text-xs font-medium;

	button {
		@apply px-2 py-2;

		&.active {
			@apply bg-I0 text-M1;
		}
	}
}

.area {
	@apply relative w-full h-48 rounded-lg cursor-crosshair touch-none select-none;

	&.hatched {
		background: repeating-linear-gradient(-45deg, rgb(from var(--M5) r g b / 0.25) 0 4px, transparent 4px 10px);
	}

	canvas {
		@apply absolute inset-0 size-full rounded-lg;
	}

	.handle {
		@apply size-5;
	}
}

.bar {
	@apply relative h-3 rounded-full cursor-pointer touch-none select-none;

	.handle {
		@apply top-1/2 size-5 overflow-hidden;
	}
}

.handle {
	@apply absolute -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-white pointer-events-none;
	box-shadow:
		0 0 0 1px rgb(0 0 0 / 0.3),
		0 1px 3px rgb(0 0 0 / 0.4);
}

/* the handles stand out by half of their size, which must still grab them */
.area::before,
.bar::before {
	content: '';
	@apply absolute -inset-2.5;
}

.area:focus-visible,
.bar:focus-visible {
	@apply outline outline-2 outline-offset-2 outline-I0;
}

.notations {
	@apply grid gap-2;
	grid-template-columns: repeat(auto-fit, minmax(9.5rem, 1fr));
}

.notation {
	@apply flex flex-col gap-0.5 text-xs;

	&.wide {
		grid-column: span 2;
	}

	span {
		@apply px-0.5 text-ST0;
	}
}
</style>
