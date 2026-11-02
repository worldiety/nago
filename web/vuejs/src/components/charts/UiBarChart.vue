<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import CartesianAxes from '@/components/charts/CartesianAxes.vue';
import ChartOverlay from '@/components/charts/ChartOverlay.vue';
import { Cartesian } from '@/components/charts/cartesian';
import {
	type Callout,
	type CalloutRow,
	type Label,
	categories,
	categoryIndex,
	containerStyle,
	formatValue,
	labelHeight,
	labelSize,
	num,
	seriesColor,
	seriesValues,
} from '@/components/charts/chart';
import { useChart } from '@/components/charts/useChart';
import { colorValue } from '@/components/shared/colors';
import type { BarChart } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: BarChart;
}>();

// maxBar caps the thickness of a bar, so that a chart of a few categories does not become a block of color.
const maxBar = 56;
// maxValueLabels is the most bars which get their value written on them.
const maxValueLabels = 100;

const { t } = useI18n();

const series = computed(() => props.ui.series?.value ?? []);
const cats = computed(() => categories(props.ui.chart, series.value));
const data = computed(() => {
	const index = categoryIndex(cats.value);
	return series.value.map((s) => seriesValues(s, cats.value, index));
});
const hasData = computed(() => data.value.some((d) => d.values.some((v) => v !== undefined)));
const horizontal = computed(() => props.ui.horizontal ?? false);
const stacked = computed(() => props.ui.stacked ?? false);
const rounding = computed(() => props.ui.chart?.labelRounding);
const names = computed(() =>
	series.value.map(
		(s, i) => s.label || (series.value.length > 1 ? t('chart.series', { n: i + 1 }) : t('chart.value'))
	)
);
const colors = computed(() => series.value.map((_, i) => seriesColor(props.ui.chart, i)));

// Goal is a marker of a target value on a single bar.
interface Goal {
	series: number;
	category: number;
	value: number;
	label: string;
	color: string;
	thickness: number;
	round: boolean;
	dashed: boolean;
}

const goals = computed<Goal[]>(() => {
	if (stacked.value) {
		return [];
	}

	const out: Goal[] = [];
	for (const m of props.ui.markers?.value ?? []) {
		const value = parseFloat(m.value ?? '');
		const s = m.seriesIndex ?? 0;
		const category = data.value[s]?.categoryOf[m.dataPointIndex ?? 0] ?? -1;
		if (!isFinite(value) || category < 0) {
			continue;
		}

		out.push({
			series: s,
			category,
			value,
			label: m.label ?? '',
			color: colorValue(m.color) || 'currentColor',
			thickness: (horizontal.value ? m.width : m.height) || 2,
			round: m.round ?? false,
			dashed: m.dashed ?? false,
		});
	}

	return out;
});

// extent returns the smallest and the largest value. Stacked bars sum the positive and the negative values of a
// category separately.
const extent = computed<[number, number]>(() => {
	let lo = 0;
	let hi = 0;
	if (stacked.value) {
		cats.value.forEach((_, c) => {
			let pos = 0;
			let neg = 0;
			for (const d of data.value) {
				const v = d.values[c] ?? 0;
				if (v >= 0) pos += v;
				else neg += v;
			}

			lo = Math.min(lo, neg);
			hi = Math.max(hi, pos);
		});

		return [lo, hi];
	}

	for (const d of data.value) {
		for (const v of d.values) {
			if (v !== undefined) {
				lo = Math.min(lo, v);
				hi = Math.max(hi, v);
			}
		}
	}

	for (const g of goals.value) {
		lo = Math.min(lo, g.value);
		hi = Math.max(hi, g.value);
	}

	return [lo, hi];
});

const { root, width, height, measure, inspected, pointer, keyboard, listeners } = useChart(
	() => (hasData.value ? cats.value.length : 0),
	(x, y) => layout.value?.categoryAt(x, y) ?? -1
);

const layout = computed(() => {
	if (width.value <= 0 || height.value <= 0 || !hasData.value) {
		return undefined;
	}

	const [lo, hi] = extent.value;
	return new Cartesian({
		width: width.value,
		height: height.value,
		categories: cats.value,
		lo,
		hi,
		includeZero: true,
		horizontal: horizontal.value,
		xTitle: props.ui.chart?.xAxisTitle,
		yTitle: props.ui.chart?.yAxisTitle,
		legend: names.value.map((label, i) => ({ label, color: colors.value[i], kind: 'square' as const })),
		toolbar: props.ui.chart?.downloadable ?? false,
		measure: measure.value,
	});
});

// Bar is a drawn bar or segment of a stack, in the coordinates of the drawing.
interface Bar {
	series: number;
	category: number;
	value: number;
	x: number;
	y: number;
	w: number;
	h: number;
	// start and end are the positions of the bar on the value axis, from the baseline outwards.
	start: number;
	end: number;
	radius: number;
}

const bars = computed<Bar[]>(() => {
	const l = layout.value;
	if (!l) {
		return [];
	}

	const groups = stacked.value ? 1 : Math.max(series.value.length, 1);
	const gap = groups > 1 ? Math.min(2, l.bandSize * 0.05) : 0;
	const thickness = Math.min((l.bandSize * 0.8 - gap * (groups - 1)) / groups, maxBar);
	const groupWidth = thickness * groups + gap * (groups - 1);
	const pos = new Array<number>(cats.value.length).fill(0);
	const neg = new Array<number>(cats.value.length).fill(0);
	const out: Bar[] = [];
	data.value.forEach((d, s) => {
		d.values.forEach((v, c) => {
			if (v === undefined) {
				return;
			}

			let from = 0;
			let to = v;
			if (stacked.value) {
				const sums = v >= 0 ? pos : neg;
				from = sums[c];
				sums[c] += v;
				to = sums[c];
			}

			const across = l.bandStart(c) + (l.bandSize - groupWidth) / 2 + (stacked.value ? 0 : s * (thickness + gap));
			let start = l.valuePos(from);
			let end = l.valuePos(to);
			if (stacked.value && Math.abs(end - start) > 2) {
				// the segments of a stack are kept a pixel apart
				const dir = Math.sign(end - start);
				start += dir * 0.5;
				end -= dir * 0.5;
			}

			const length = Math.abs(end - start);
			const along = Math.min(start, end);
			const radius = Math.min(3, thickness / 4, length / 2);
			if (horizontal.value) {
				out.push({
					series: s,
					category: c,
					value: v,
					x: along,
					y: across,
					w: length,
					h: thickness,
					start,
					end,
					radius,
				});
			} else {
				out.push({
					series: s,
					category: c,
					value: v,
					x: across,
					y: along,
					w: thickness,
					h: length,
					start,
					end,
					radius,
				});
			}
		});
	});

	return out;
});

interface ValueLabel extends Label {
	inside: boolean;
}

// valueLabels writes the value of each bar into its segment when stacked, otherwise at its end, and leaves out a
// label that does not fit or would run into the one before.
const valueLabels = computed<ValueLabel[]>(() => {
	const l = layout.value;
	if (!l || bars.value.length > maxValueLabels) {
		return [];
	}

	const p = l.plot;
	const out: ValueLabel[] = [];
	let edge = -Infinity;
	const sorted = [...bars.value].sort((a, b) => (horizontal.value ? a.y - b.y : a.x - b.x));
	for (const b of sorted) {
		if (b.w < 0.5 && b.h < 0.5) {
			continue;
		}

		const text = formatValue(b.value, rounding.value);
		const w = measure.value.width(text);
		const fitsInside = w <= b.w - 4 && labelHeight <= b.h + (horizontal.value ? 2 : -2);
		const cx = b.x + b.w / 2;
		const cy = b.y + b.h / 2;
		if (stacked.value) {
			if (fitsInside) {
				out.push({ x: cx, y: cy, text, anchor: 'middle', inside: true });
			}

			continue;
		}

		const outward = Math.sign(b.end - b.start) || 1;
		if (horizontal.value) {
			const x = b.end + outward * 4;
			const fits = outward > 0 ? x + w <= p.x + p.w + 4 : x - w >= p.x - 4;
			if (cy - labelHeight / 2 < edge || labelHeight > b.h + 4) {
				continue;
			}

			if (fits) {
				out.push({ x, y: cy, text, anchor: outward > 0 ? 'start' : 'end', inside: false });
			} else if (fitsInside) {
				out.push({ x: b.end - outward * 4, y: cy, text, anchor: outward > 0 ? 'end' : 'start', inside: true });
			} else {
				continue;
			}

			edge = cy + labelHeight / 2;
			continue;
		}

		const y = b.end + outward * (labelHeight / 2 + 1);
		const fits = y - labelHeight / 2 >= p.y - 2 && y + labelHeight / 2 <= p.y + p.h + 2;
		if (cx - w / 2 < edge + 2) {
			continue;
		}

		if (fits) {
			out.push({ x: cx, y, text, anchor: 'middle', inside: false });
		} else if (fitsInside) {
			out.push({ x: cx, y: b.end - outward * (labelHeight / 2 + 2), text, anchor: 'middle', inside: true });
		} else {
			continue;
		}

		edge = cx + w / 2;
	}

	return out;
});

const goalMarks = computed(() => {
	const l = layout.value;
	if (!l) {
		return [];
	}

	return goals.value.flatMap((g) => {
		const b = bars.value.find((b) => b.series === g.series && b.category === g.category);
		if (!b) {
			return [];
		}

		const pos = l.valuePos(Math.min(Math.max(g.value, l.min), l.max));
		const center = horizontal.value ? b.y + b.h / 2 : b.x + b.w / 2;
		const half = (horizontal.value ? b.h : b.w) / 2;
		return [
			{
				goal: g,
				cx: horizontal.value ? pos : center,
				cy: horizontal.value ? center : pos,
				x1: horizontal.value ? pos : center - half,
				y1: horizontal.value ? center - half : pos,
				x2: horizontal.value ? pos : center + half,
				y2: horizontal.value ? center + half : pos,
			},
		];
	});
});

const band = computed(() => {
	const l = layout.value;
	const i = inspected.value;
	if (!l || i < 0 || i >= cats.value.length) {
		return undefined;
	}

	const p = l.plot;
	const start = l.bandStart(i);
	return horizontal.value
		? { x: p.x, y: start, width: p.w, height: l.bandSize }
		: { x: start, y: p.y, width: l.bandSize, height: p.h };
});

const callout = computed<Callout | undefined>(() => {
	const l = layout.value;
	const i = inspected.value;
	if (!l || i < 0 || i >= cats.value.length) {
		return undefined;
	}

	const rows: CalloutRow[] = [];
	data.value.forEach((d, s) => {
		const v = d.values[i];
		if (v !== undefined) {
			rows.push({
				label: names.value[s],
				value: formatValue(v, rounding.value),
				color: colors.value[s],
				kind: 'square',
			});
		}
	});

	for (const g of goals.value) {
		if (g.category === i) {
			rows.push({
				label: g.label || t('chart.value'),
				value: formatValue(g.value, rounding.value),
				color: g.color,
				kind: g.round ? 'dot' : 'line',
			});
		}
	}

	if (horizontal.value) {
		return {
			title: cats.value[i],
			rows,
			x: pointer.value?.x ?? l.plot.x + l.plot.w / 2,
			y: pointer.value?.y ?? l.bandCenter(i),
			top: 0,
		};
	}

	// beside the band, not over it
	return { title: cats.value[i], rows, x: l.bandCenter(i), half: l.bandSize / 2 - 6, top: l.plot.y + 4 };
});

const announcement = computed(() => {
	const c = callout.value;
	if (!keyboard.value || !c) {
		return '';
	}

	return `${c.title}: ${c.rows.map((r) => `${r.label} ${r.value}`).join(', ')}`;
});

const ariaLabel = computed(() => {
	const named = series.value.map((s) => s.label).filter((s) => !!s);
	return named.length > 0 ? `${t('chart.barChart')}: ${named.join(', ')}` : t('chart.barChart');
});

function table(): string[][] {
	const rows = [[props.ui.chart?.xAxisTitle || t('chart.category'), ...names.value]];
	cats.value.forEach((c, i) =>
		rows.push([c, ...data.value.map((d) => (d.values[i] === undefined ? '' : String(d.values[i])))])
	);
	return rows;
}
</script>

<template>
	<div
		ref="root"
		class="chart"
		:style="containerStyle(ui.chart)"
		tabindex="0"
		role="figure"
		:aria-label="ariaLabel"
		v-on="listeners"
	>
		<svg
			v-if="layout"
			class="chart-drawing"
			:width="width"
			:height="height"
			:viewBox="`0 0 ${width} ${height}`"
			aria-hidden="true"
		>
			<rect v-if="band" v-bind="band" fill="currentColor" fill-opacity="0.06" data-export="false" />
			<CartesianAxes :layout="layout" />
			<g>
				<template v-for="(b, i) in bars" :key="i">
					<rect
						v-if="b.w >= 0.5 && b.h >= 0.5"
						:x="num(b.x)"
						:y="num(b.y)"
						:width="num(b.w)"
						:height="num(b.h)"
						:rx="num(b.radius)"
						:fill="colors[b.series]"
					/>
				</template>
			</g>
			<g>
				<template v-for="(m, i) in goalMarks" :key="i">
					<circle v-if="m.goal.round" :cx="m.cx" :cy="m.cy" :r="m.goal.thickness / 2" :fill="m.goal.color" />
					<line
						v-else
						:x1="m.x1"
						:y1="m.y1"
						:x2="m.x2"
						:y2="m.y2"
						:stroke="m.goal.color"
						:stroke-width="m.goal.thickness"
						:stroke-dasharray="m.goal.dashed ? '3 3' : undefined"
					/>
				</template>
			</g>
			<g :font-size="labelSize">
				<text
					v-for="(v, i) in valueLabels"
					:key="i"
					:x="num(v.x)"
					:y="num(v.y)"
					:text-anchor="v.anchor"
					dominant-baseline="central"
					:fill="v.inside ? '#ffffff' : 'currentColor'"
					:fill-opacity="v.inside ? 1 : 0.75"
					:stroke="v.inside ? 'rgba(0,0,0,0.35)' : undefined"
					:stroke-width="v.inside ? 2 : undefined"
					:paint-order="v.inside ? 'stroke' : undefined"
					:font-weight="v.inside ? 500 : undefined"
				>
					{{ v.text }}
				</text>
			</g>
		</svg>
		<ChartOverlay
			:width="width"
			:height="height"
			:callout="callout"
			:announcement="announcement"
			:message="hasData ? undefined : ui.chart?.noDataMessage || t('chart.noData')"
			:downloadable="ui.chart?.downloadable ?? false"
			:table="table"
		/>
	</div>
</template>
