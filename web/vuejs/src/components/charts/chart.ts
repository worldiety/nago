/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { activeLocale } from '@/i18n';
import type { Chart as ChartProto, ChartSeries } from '@/shared/proto/nprotoc_gen';
import { RoundingTypeValues } from '@/shared/proto/nprotoc_gen';

// labelSize is the font size of the axis labels, the legend and the value labels, titleSize the one of the titles.
export const labelSize = 11;
export const titleSize = 12;
export const labelHeight = Math.ceil(labelSize * 1.3);
export const titleHeight = Math.ceil(titleSize * 1.3);
// pad is the gap between a label and what it labels.
export const pad = 6;
// toolbarSize is the room the download button takes in the top trailing corner.
export const toolbarSize = 28;
// paletteSize is the number of default series colors, see charts.css.
const paletteSize = 8;

export interface Rect {
	x: number;
	y: number;
	w: number;
	h: number;
}

// Label is a text of the drawing, anchored at its start, middle or end and vertically centered on y.
export interface Label {
	x: number;
	y: number;
	text: string;
	anchor: 'start' | 'middle' | 'end';
}

export type SwatchKind = 'dot' | 'square' | 'line';

export interface LegendEntry {
	label: string;
	color: string;
	kind: SwatchKind;
}

export interface PlacedLegendEntry extends LegendEntry {
	x: number;
	y: number;
}

export interface CalloutRow {
	label: string;
	value: string;
	color: string;
	kind: SwatchKind;
	detail?: string;
}

// Callout is the box of an inspection. It stands beside x, at the top of the plot or, given y, beside the
// pointer.
export interface Callout {
	title: string;
	rows: CalloutRow[];
	x: number;
	// half is half the width of what the callout stands beside, e.g. a band of a bar chart.
	half?: number;
	y?: number;
	top: number;
}

// --- numbers -----------------------------------------------------------------

const formats = new Map<string, Intl.NumberFormat>();

function numberFormat(options: Intl.NumberFormatOptions): Intl.NumberFormat {
	const key = JSON.stringify(options);
	let f = formats.get(key);
	if (!f) {
		try {
			f = new Intl.NumberFormat(activeLocale, options);
		} catch {
			f = new Intl.NumberFormat(undefined, options);
		}
		formats.set(key, f);
	}

	return f;
}

export function round(v: number, rounding?: number): number {
	switch (rounding) {
		case RoundingTypeValues.Round:
			return Math.round(v);
		case RoundingTypeValues.Floor:
			return Math.floor(v);
		case RoundingTypeValues.Ceiling:
			return Math.ceil(v);
	}

	return v;
}

// formatValue formats a single value as the labels and the inspection show it: rounded as the chart says, or
// with four significant digits, more when they are whole.
export function formatValue(v: number, rounding?: number): string {
	v = round(v, rounding);
	const a = Math.abs(v);
	if (a === 0) {
		return '0';
	}

	if (a >= 1e6) {
		return numberFormat({ maximumFractionDigits: 0 }).format(v);
	}

	if (a < 1e-3) {
		return numberFormat({ maximumSignificantDigits: 4 }).format(v);
	}

	const decimals = Math.max(0, 3 - Math.floor(Math.log10(a)));
	return numberFormat({ maximumFractionDigits: decimals }).format(v);
}

// formatPercent formats a share of 0..1 as percent, rounded as the chart says or to whole percents.
export function formatPercent(share: number, rounding?: number): string {
	const raw = share * 100;
	if (raw > 0 && raw < 1 && !rounding) {
		return '<' + numberFormat({ style: 'percent' }).format(0.01);
	}

	const p = round(raw, rounding || RoundingTypeValues.Round);
	return numberFormat({ style: 'percent', maximumFractionDigits: 0 }).format(p / 100);
}

export interface Tick {
	value: number;
	label: string;
}

// niceStep returns the step of 1, 2 or 5 times a power of ten that divides span into about count parts, but
// never more.
export function niceStep(span: number, count: number): number {
	if (!(span > 0) || !isFinite(span) || count < 1) {
		return 1;
	}

	const raw = span / count;
	const p = Math.pow(10, Math.floor(Math.log10(raw)));
	for (const m of [1, 2, 5, 10]) {
		if (m * p >= raw) {
			return m * p;
		}
	}

	return 10 * p;
}

// niceDomain widens lo..hi to multiples of the step its ticks would have and makes an empty one a range.
export function niceDomain(lo: number, hi: number, count: number): [number, number] {
	if (lo === hi) {
		const d = Math.max(Math.abs(lo) * 0.1, 1);
		lo = lo === 0 ? 0 : lo - d;
		hi = hi + d;
	}

	const step = niceStep(hi - lo, Math.max(count, 1));
	return [Math.floor(lo / step + 1e-9) * step, Math.ceil(hi / step - 1e-9) * step];
}

// numberTicks returns the multiples of a nice step in lo..hi, about count of them.
export function numberTicks(lo: number, hi: number, count: number): Tick[] {
	const step = niceStep(hi - lo, Math.max(count, 1));
	const format = tickFormat(step);
	const ticks: Tick[] = [];
	for (let k = Math.ceil(lo / step - 1e-9); ticks.length <= 1000; k++) {
		// multiples are rounded, so that 3*0.1 does not become 0.30000000000000004
		const v = Number((k * step).toPrecision(12)) || 0;
		if (v > hi + step * 1e-9) {
			break;
		}

		ticks.push({ value: v, label: format(v) });
	}

	return ticks;
}

// tickFormat returns a format with as many decimals as the step needs, and thousands as k and millions as M
// from ten thousand on, as a compact axis label wants them.
function tickFormat(step: number): (v: number) => string {
	const decimals = step > 0 && step < 1 ? Math.ceil(-Math.log10(step) - 1e-9) : 0;
	return (v: number) => {
		if (v === 0) {
			return '0';
		}

		const a = Math.abs(v);
		for (const [div, suffix] of [
			[1e9, 'G'],
			[1e6, 'M'],
			[1e3, 'k'],
		] as const) {
			if (a >= 10 * div || (a >= div && step >= div / 10)) {
				const q = step / div;
				const d = q < 1 ? Math.ceil(-Math.log10(q) - 1e-9) : 0;
				return numberFormat({ maximumFractionDigits: d }).format(v / div) + suffix;
			}
		}

		return numberFormat({ minimumFractionDigits: decimals, maximumFractionDigits: decimals }).format(v);
	};
}

// --- data --------------------------------------------------------------------

// categories returns the labels of the category axis. The labels of the chart win, otherwise the X values of
// all series are collected in the order of their first appearance.
export function categories(chart: ChartProto | undefined, series: ChartSeries[]): string[] {
	const labels = chart?.labels?.value ?? [];
	if (labels.length > 0) {
		return labels;
	}

	const cats: string[] = [];
	const seen = new Set<string>();
	for (const s of series) {
		for (const dp of s.dataPoints?.value ?? []) {
			const x = dp.x ?? '';
			if (!seen.has(x)) {
				seen.add(x);
				cats.push(x);
			}
		}
	}

	return cats;
}

export interface SeriesValues {
	// values holds the value of each category, undefined where the series has none.
	values: (number | undefined)[];
	// categoryOf holds the category of each data point, -1 if it has none.
	categoryOf: number[];
}

// seriesValues assigns the data points of the series to the categories. A data point whose X value is not a
// category is placed by its position within the series.
export function seriesValues(series: ChartSeries, cats: string[], index: Map<string, number>): SeriesValues {
	const values: (number | undefined)[] = new Array(cats.length).fill(undefined);
	const categoryOf: number[] = [];
	(series.dataPoints?.value ?? []).forEach((dp, i) => {
		const c = index.get(dp.x ?? '') ?? i;
		if (c >= cats.length) {
			categoryOf.push(-1);
			return;
		}

		values[c] = dp.y ?? 0;
		categoryOf.push(c);
	});

	return { values, categoryOf };
}

export function categoryIndex(cats: string[]): Map<string, number> {
	const index = new Map<string, number>();
	cats.forEach((c, i) => {
		if (!index.has(c)) {
			index.set(c, i);
		}
	});

	return index;
}

// seriesColor returns the color of the series or slice with the given index: the colors of the chart in turn,
// or the default palette.
export function seriesColor(chart: ChartProto | undefined, i: number): string {
	const colors = (chart?.colors?.value ?? []).map((c) => colorValue(c)).filter((c) => c.length > 0);
	if (colors.length > 0) {
		return colors[i % colors.length];
	}

	return `var(--chart-color-${i % paletteSize})`;
}

// containerStyle returns the frame of the chart. Without a width it takes the width it is given, without a height
// it keeps an aspect ratio up to the default height of the server side svgchart package.
export function containerStyle(chart: ChartProto | undefined): string {
	const frame = chart?.frame;
	const styles = frameCSS(frame);
	if (!frame?.width) {
		styles.push('width:100%');
	}

	if (!frame?.height) {
		styles.push('aspect-ratio:16/9');
		if (!frame?.maxHeight) {
			styles.push('max-height:20rem');
		}
	}

	return styles.join(';');
}

// --- text --------------------------------------------------------------------

let canvas: CanvasRenderingContext2D | null | undefined;

// TextMeasurer measures the width of texts in the font of the chart.
export class TextMeasurer {
	private readonly cache = new Map<string, number>();

	constructor(private readonly family: string) {}

	width(text: string, size: number = labelSize, weight: number = 400): number {
		const key = `${size}|${weight}|${text}`;
		let w = this.cache.get(key);
		if (w !== undefined) {
			return w;
		}

		if (canvas === undefined) {
			canvas = document.createElement('canvas').getContext('2d');
		}

		if (canvas) {
			canvas.font = `${weight} ${size}px ${this.family}`;
			w = canvas.measureText(text).width;
		} else {
			w = text.length * size * 0.58;
		}

		if (this.cache.size > 5000) {
			this.cache.clear();
		}

		this.cache.set(key, w);
		return w;
	}

	// truncate shortens the text with an ellipsis, so that it is at most width wide.
	truncate(text: string, width: number, size: number = labelSize): string {
		if (this.width(text, size) <= width) {
			return text;
		}

		let lo = 0;
		let hi = text.length;
		while (lo < hi) {
			const mid = Math.ceil((lo + hi) / 2);
			if (this.width(text.slice(0, mid) + '…', size) <= width) {
				lo = mid;
			} else {
				hi = mid - 1;
			}
		}

		return lo > 0 ? text.slice(0, lo) + '…' : '…';
	}
}

// --- legend ------------------------------------------------------------------

// legendSwatch is the room of the swatch in front of a legend entry.
export const legendSwatch = 14;

// flowLegend flows the entries into rows of the given width and returns them with their positions from the top
// of the legend and its height.
export function flowLegend(
	entries: LegendEntry[],
	width: number,
	measure: TextMeasurer
): { items: PlacedLegendEntry[]; height: number } {
	const items: PlacedLegendEntry[] = [];
	let x = 0;
	let y = 0;
	for (const e of entries) {
		const label = measure.truncate(e.label, Math.max(width - legendSwatch, 20));
		const w = legendSwatch + measure.width(label);
		if (x > 0 && x + w > width) {
			x = 0;
			y += labelHeight + 4;
		}

		items.push({ ...e, label, x, y });
		x += w + 14;
	}

	return { items, height: items.length > 0 ? y + labelHeight : 0 };
}

// --- paths -------------------------------------------------------------------

export interface Point {
	x: number;
	y: number;
}

// num rounds a coordinate to two decimals, which keeps the markup small.
export function num(v: number): number {
	if (!isFinite(v)) {
		return 0;
	}

	return Math.round(v * 100) / 100;
}

export const CurveStraight = 0;
export const CurveSmooth = 1;
export const CurveStepline = 2;

// linePath returns the path through the points: straight, as a monotone cubic spline, which never overshoots the
// values, or as steps which hold each value until the next point.
export function linePath(points: Point[], curve?: number): string {
	if (points.length === 0) {
		return '';
	}

	let d = `M${num(points[0].x)} ${num(points[0].y)}`;
	switch (curve) {
		case CurveStepline:
			for (let i = 1; i < points.length; i++) {
				d += `H${num(points[i].x)}V${num(points[i].y)}`;
			}
			break;
		case CurveSmooth: {
			const t = monotoneTangents(points);
			for (let i = 1; i < points.length; i++) {
				const p0 = points[i - 1];
				const p1 = points[i];
				const dx = (p1.x - p0.x) / 3;
				d += `C${num(p0.x + dx)} ${num(p0.y + t[i - 1] * dx)} ${num(p1.x - dx)} ${num(p1.y - t[i] * dx)} ${num(p1.x)} ${num(p1.y)}`;
			}
			break;
		}
		default:
			for (let i = 1; i < points.length; i++) {
				d += `L${num(points[i].x)} ${num(points[i].y)}`;
			}
	}

	return d;
}

// areaPath returns the path of the area between the line through the points and the baseline.
export function areaPath(points: Point[], curve: number | undefined, base: number): string {
	if (points.length < 2) {
		return '';
	}

	const last = points[points.length - 1];
	return `${linePath(points, curve)}L${num(last.x)} ${num(base)}L${num(points[0].x)} ${num(base)}Z`;
}

// monotoneTangents returns the slope at each point following Fritsch and Carlson, so that the interpolated curve
// does not overshoot the data.
function monotoneTangents(points: Point[]): number[] {
	const n = points.length;
	const tangents = new Array<number>(n).fill(0);
	if (n < 2) {
		return tangents;
	}

	const slopes: number[] = [];
	for (let i = 0; i < n - 1; i++) {
		const dx = points[i + 1].x - points[i].x;
		slopes.push(dx === 0 ? 0 : (points[i + 1].y - points[i].y) / dx);
	}

	tangents[0] = slopes[0];
	tangents[n - 1] = slopes[n - 2];
	for (let i = 1; i < n - 1; i++) {
		tangents[i] = slopes[i - 1] * slopes[i] <= 0 ? 0 : (slopes[i - 1] + slopes[i]) / 2;
	}

	for (let i = 0; i < n - 1; i++) {
		if (slopes[i] === 0) {
			tangents[i] = 0;
			tangents[i + 1] = 0;
			continue;
		}

		const a = tangents[i] / slopes[i];
		const b = tangents[i + 1] / slopes[i];
		const s = a * a + b * b;
		if (s > 9) {
			const t = 3 / Math.sqrt(s);
			tangents[i] = t * a * slopes[i];
			tangents[i + 1] = t * b * slopes[i];
		}
	}

	return tangents;
}

// m4 reduces points which are more than the pixels can show: of the points of a pixel column it keeps the first
// and the last, the smallest and the largest, which draws the same pixels as all of them.
export function m4(points: Point[]): Point[] {
	const out: Point[] = [];
	let col = NaN;
	let group: Point[] = [];
	const flush = () => {
		if (group.length <= 4) {
			out.push(...group);
			return;
		}

		let lo = 0;
		let hi = 0;
		group.forEach((p, i) => {
			if (p.y < group[lo].y) lo = i;
			if (p.y > group[hi].y) hi = i;
		});
		const keep = new Set([0, lo, hi, group.length - 1]);
		[...keep].sort((a, b) => a - b).forEach((i) => out.push(group[i]));
	};

	for (const p of points) {
		const c = Math.floor(p.x);
		if (c !== col) {
			flush();
			group = [];
			col = c;
		}

		group.push(p);
	}

	flush();
	return out;
}

// polar returns the point at radius r and angle a, clockwise from twelve o'clock, about c.
export function polar(cx: number, cy: number, r: number, a: number): Point {
	return { x: cx + r * Math.sin(a), y: cy - r * Math.cos(a) };
}

// sectorPath returns the path of the slice from a0 to a1, clockwise from twelve o'clock, between the radii r0 and
// r1 about c. Its sides are moved inwards by half of the gap, so that two slices are gap pixels apart all along.
export function sectorPath(
	cx: number,
	cy: number,
	r0: number,
	r1: number,
	a0: number,
	a1: number,
	gap: number
): string {
	const sweep = a1 - a0;
	const p = (pt: Point) => `${num(pt.x)} ${num(pt.y)}`;
	if (sweep >= 2 * Math.PI - 1e-6) {
		// a single arc cannot describe a full circle, so it takes two halves
		const outer =
			`M${p(polar(cx, cy, r1, 0))}A${num(r1)} ${num(r1)} 0 1 1 ${p(polar(cx, cy, r1, Math.PI))}` +
			`A${num(r1)} ${num(r1)} 0 1 1 ${p(polar(cx, cy, r1, 0))}Z`;
		if (r0 <= 0) {
			return outer;
		}

		return (
			outer +
			`M${p(polar(cx, cy, r0, 0))}A${num(r0)} ${num(r0)} 0 1 0 ${p(polar(cx, cy, r0, Math.PI))}` +
			`A${num(r0)} ${num(r0)} 0 1 0 ${p(polar(cx, cy, r0, 0))}Z`
		);
	}

	// at radius r the sides move by the angle of half the gap
	const inset = (r: number) => (r > 0 ? Math.min(Math.asin(Math.min(gap / 2 / r, 1)), (sweep / 2) * 0.95) : 0);
	const o = inset(r1);
	const large = (b: number, a: number) => (b - a > Math.PI ? 1 : 0);
	let d = `M${p(polar(cx, cy, r1, a0 + o))}A${num(r1)} ${num(r1)} 0 ${large(a1 - o, a0 + o)} 1 ${p(polar(cx, cy, r1, a1 - o))}`;
	if (r0 > 0) {
		const i = inset(r0);
		d += `L${p(polar(cx, cy, r0, a1 - i))}A${num(r0)} ${num(r0)} 0 ${large(a1 - i, a0 + i)} 0 ${p(polar(cx, cy, r0, a0 + i))}Z`;
		return d;
	}

	// the tip of a pie slice moves out along its middle
	const tip = gap > 0 ? Math.min(gap / 2 / Math.max(Math.sin(sweep / 2), 0.05), r1 * 0.3) : 0;
	return d + `L${p(polar(cx, cy, tip, (a0 + a1) / 2))}Z`;
}
