<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import ChartLegend from '@/components/charts/ChartLegend.vue';
import ChartOverlay from '@/components/charts/ChartOverlay.vue';
import {
	type Callout,
	type Label,
	type LegendEntry,
	type PlacedLegendEntry,
	containerStyle,
	flowLegend,
	formatPercent,
	formatValue,
	labelHeight,
	legendSwatch,
	num,
	pad,
	polar,
	sectorPath,
	seriesColor,
	titleSize,
	toolbarSize,
} from '@/components/charts/chart';
import { useChart } from '@/components/charts/useChart';
import type { PieChart } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: PieChart;
}>();

// pop is how far the inspected slice stands out of the circle.
const pop = 6;
// gap is the distance between two slices.
const gap = 2;

const { t } = useI18n();

const rounding = computed(() => props.ui.chart?.labelRounding);

interface Slice {
	label: string;
	value: number;
	share: number;
	color: string;
	// a0 and a1 are the angles, clockwise from twelve o'clock.
	a0: number;
	a1: number;
}

// slices are the positive data points of the first series. The labels of the chart win over the X values.
const slices = computed<Slice[]>(() => {
	const dps = props.ui.series?.value[0]?.dataPoints?.value ?? [];
	const labels = props.ui.chart?.labels?.value ?? [];
	const total = dps.reduce((sum, dp) => sum + Math.max(dp.y ?? 0, 0), 0);
	if (!(total > 0)) {
		return [];
	}

	let at = 0;
	return dps.flatMap((dp, i) => {
		const value = dp.y ?? 0;
		if (!(value > 0)) {
			return [];
		}

		const share = value / total;
		const a0 = at * 2 * Math.PI;
		at += share;
		return [
			{
				label: labels[i] || dp.x || '',
				value,
				share,
				color: seriesColor(props.ui.chart, i),
				a0,
				a1: at * 2 * Math.PI,
			},
		];
	});
});

const { root, width, height, measure, inspected, pointer, keyboard, listeners } = useChart(
	() => slices.value.length,
	(x, y) => sliceAt(x, y)
);

// layout puts the circle as large as the room leaves it, with the legend beside it when the chart is wide and
// below it otherwise.
const layout = computed(() => {
	if (width.value <= 0 || height.value <= 0 || slices.value.length === 0) {
		return undefined;
	}

	const top = props.ui.chart?.downloadable ? toolbarSize + 4 : 0;
	const entries: LegendEntry[] = slices.value.map((s) => ({ label: s.label, color: s.color, kind: 'square' }));
	const labelWidth = Math.min(Math.max(...entries.map((e) => measure.value.width(e.label))), width.value * 0.4);
	const legendWidth = labelWidth + legendSwatch;
	const legendHeight = entries.length * (labelHeight + 4) - 4;
	const area = { x: 0, y: top, w: width.value, h: height.value - top };
	let legend: PlacedLegendEntry[];
	if (width.value - legendWidth - 2 * pad >= area.h) {
		area.w = width.value - legendWidth - 2 * pad;
		const y = top + Math.max((area.h - legendHeight) / 2, 0);
		legend = entries.map((e, i) => ({
			...e,
			label: measure.value.truncate(e.label, labelWidth),
			x: area.w + 2 * pad,
			y: y + i * (labelHeight + 4),
		}));
	} else {
		const flow = flowLegend(entries, width.value, measure.value);
		area.h = Math.max(area.h - flow.height - pad * 2, 1);
		legend = flow.items.map((e) => ({ ...e, y: e.y + height.value - flow.height }));
	}

	const r = Math.max(Math.min(area.w, area.h) / 2 - pop, 1);
	return {
		cx: area.x + area.w / 2,
		cy: area.y + area.h / 2,
		r,
		inner: props.ui.showAsDonut ? r * 0.6 : 0,
		legend,
	};
});

// sliceAt returns the slice at the position, or -1 outside of the circle.
function sliceAt(x: number, y: number): number {
	const l = layout.value;
	if (!l) {
		return -1;
	}

	const dx = x - l.cx;
	const dy = y - l.cy;
	const d = Math.hypot(dx, dy);
	let a = Math.atan2(dx, -dy);
	if (a < 0) {
		a += 2 * Math.PI;
	}

	if (d > l.r + pop || d < l.inner) {
		return -1;
	}

	return slices.value.findIndex((s) => a >= s.a0 && a < s.a1);
}

const paths = computed(() => {
	const l = layout.value;
	if (!l) {
		return [];
	}

	const g = slices.value.length > 1 ? gap : 0;
	return slices.value.map((s, i) => {
		const mid = (s.a0 + s.a1) / 2;
		const selected = inspected.value === i;
		return {
			d: sectorPath(l.cx, l.cy, l.inner, l.r, s.a0, s.a1, g),
			color: s.color,
			// the inspected slice stands out of the circle and the others fade
			style: {
				transform: selected ? `translate(${pop * Math.sin(mid)}px, ${-pop * Math.cos(mid)}px)` : undefined,
				opacity: inspected.value >= 0 && !selected ? 0.4 : 1,
			},
		};
	});
});

// labels writes the share or the value into the slices which have room for it.
const labels = computed<(Label & { opacity: number })[]>(() => {
	const l = layout.value;
	if (!l || !props.ui.showDataLabels) {
		return [];
	}

	return slices.value.flatMap((s, i) => {
		const text = props.ui.showAbsoluteValues
			? formatValue(s.value, rounding.value)
			: formatPercent(s.share, rounding.value);
		const mid = (s.a0 + s.a1) / 2;
		// in a pie a little beyond the middle of the radius, in a ring in the middle of the ring
		const rr = l.inner > 0 ? (l.r + l.inner) / 2 : l.r * 0.62;
		const w = measure.value.width(text, titleSize, 600);
		if (w > (s.a1 - s.a0) * rr * 1.4 || labelHeight > (l.r - l.inner) * 0.9) {
			return [];
		}

		const out = inspected.value === i ? pop : 0;
		const at = polar(l.cx, l.cy, rr + out, mid);
		const opacity = inspected.value >= 0 && inspected.value !== i ? 0.4 : 1;
		return [{ x: at.x, y: at.y, text, anchor: 'middle' as const, opacity }];
	});
});

const callout = computed<Callout | undefined>(() => {
	const l = layout.value;
	const i = inspected.value;
	if (!l || i < 0 || i >= slices.value.length) {
		return undefined;
	}

	const s = slices.value[i];
	const at = pointer.value ?? polar(l.cx, l.cy, (l.r + l.inner) / 2, (s.a0 + s.a1) / 2);
	return {
		title: s.label,
		rows: [
			{
				label: props.ui.series?.value[0]?.label || t('chart.value'),
				value: formatValue(s.value, rounding.value),
				detail: formatPercent(s.share, rounding.value),
				color: s.color,
				kind: 'square',
			},
		],
		x: at.x,
		y: at.y,
		top: 0,
	};
});

const announcement = computed(() => {
	const c = callout.value;
	if (!keyboard.value || !c) {
		return '';
	}

	return `${c.title}: ${c.rows.map((r) => `${r.value} (${r.detail})`).join(', ')}`;
});

const ariaLabel = computed(() => {
	const type = props.ui.showAsDonut ? t('chart.donutChart') : t('chart.pieChart');
	const named = slices.value.map((s) => s.label).filter((s) => !!s);
	return named.length > 0 ? `${type}: ${named.join(', ')}` : type;
});

function table(): string[][] {
	const s = props.ui.series?.value[0];
	const labels = props.ui.chart?.labels?.value ?? [];
	const rows = [[t('chart.category'), s?.label || t('chart.value')]];
	(s?.dataPoints?.value ?? []).forEach((dp, i) => rows.push([labels[i] || dp.x || '', String(dp.y ?? 0)]));
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
			<path v-for="(p, i) in paths" :key="i" class="chart-slice" :d="p.d" :fill="p.color" :style="p.style" />
			<g :font-size="titleSize" font-weight="600" fill="#ffffff" stroke="rgba(0,0,0,0.35)" stroke-width="2">
				<text
					v-for="(l, i) in labels"
					:key="i"
					:x="num(l.x)"
					:y="num(l.y)"
					text-anchor="middle"
					dominant-baseline="central"
					paint-order="stroke"
					class="chart-slice"
					:style="{ opacity: l.opacity }"
				>
					{{ l.text }}
				</text>
			</g>
			<ChartLegend :items="layout.legend" />
		</svg>
		<ChartOverlay
			:width="width"
			:height="height"
			:callout="callout"
			:announcement="announcement"
			:message="slices.length > 0 ? undefined : ui.chart?.noDataMessage || t('chart.noData')"
			:downloadable="ui.chart?.downloadable ?? false"
			:table="table"
		/>
	</div>
</template>

<style scoped>
.chart-slice {
	transition:
		transform 150ms ease,
		opacity 150ms ease;
}
</style>
