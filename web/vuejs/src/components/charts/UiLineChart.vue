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
	type Point,
	type SwatchKind,
	areaPath,
	categories,
	categoryIndex,
	containerStyle,
	formatValue,
	labelHeight,
	labelSize,
	linePath,
	m4,
	num,
	seriesColor,
	seriesValues,
} from '@/components/charts/chart';
import { useChart } from '@/components/charts/useChart';
import { colorValue } from '@/components/shared/colors';
import { randomStr } from '@/components/shared/util';
import type { LineChart } from '@/shared/proto/nprotoc_gen';
import { ChartSeriesTypeValues } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: LineChart;
}>();

// maxLabels is the most values a series gets written beside its points.
const maxLabels = 50;

const { t } = useI18n();
const id = randomStr(8);

const series = computed(() => props.ui.series?.value ?? []);
const cats = computed(() => categories(props.ui.chart, series.value));
const data = computed(() => {
	const index = categoryIndex(cats.value);
	return series.value.map((s) => seriesValues(s, cats.value, index));
});
const hasData = computed(() => data.value.some((d) => d.values.some((v) => v !== undefined)));
const types = computed(() => series.value.map((s) => s.type ?? ChartSeriesTypeValues.ChartSeriesTypeLine));
const rounding = computed(() => props.ui.chart?.labelRounding);
const names = computed(() =>
	series.value.map(
		(s, i) => s.label || (series.value.length > 1 ? t('chart.series', { n: i + 1 }) : t('chart.value'))
	)
);
const colors = computed(() => series.value.map((_, i) => seriesColor(props.ui.chart, i)));
const kinds = computed<SwatchKind[]>(() =>
	types.value.map((type) => (type === ChartSeriesTypeValues.ChartSeriesTypeLine ? 'dot' : 'square'))
);
const columnSeries = computed(() =>
	series.value.map((_, i) => i).filter((i) => types.value[i] === ChartSeriesTypeValues.ChartSeriesTypeColumn)
);
const markerSize = computed(() => props.ui.markers?.size ?? 0);
const markerStroke = computed(() => colorValue(props.ui.markers?.borderColor) || 'var(--M1)');

const { root, width, height, measure, inspected, keyboard, listeners } = useChart(
	() => (hasData.value ? cats.value.length : 0),
	(x, y) => layout.value?.categoryAt(x, y) ?? -1
);

const layout = computed(() => {
	if (width.value <= 0 || height.value <= 0 || !hasData.value) {
		return undefined;
	}

	let lo = Infinity;
	let hi = -Infinity;
	for (const d of data.value) {
		for (const v of d.values) {
			if (v !== undefined) {
				lo = Math.min(lo, v);
				hi = Math.max(hi, v);
			}
		}
	}

	return new Cartesian({
		width: width.value,
		height: height.value,
		categories: cats.value,
		lo,
		hi,
		// bars and areas grow from zero, lines fit the values
		includeZero: types.value.some((type) => type !== ChartSeriesTypeValues.ChartSeriesTypeLine),
		horizontal: false,
		xTitle: props.ui.chart?.xAxisTitle,
		yTitle: props.ui.chart?.yAxisTitle,
		legend: names.value.map((label, i) => ({ label, color: colors.value[i], kind: kinds.value[i] })),
		toolbar: props.ui.chart?.downloadable ?? false,
		measure: measure.value,
	});
});

interface Column {
	series: number;
	value: number;
	x: number;
	y: number;
	w: number;
	h: number;
	end: number;
}

// columns are the series of type column, drawn as grouped bars behind the lines.
const columns = computed<Column[]>(() => {
	const l = layout.value;
	const n = columnSeries.value.length;
	if (!l || n === 0) {
		return [];
	}

	const gap = n > 1 ? Math.min(2, l.bandSize * 0.05) : 0;
	const thickness = Math.min((l.bandSize * 0.8 - gap * (n - 1)) / n, 56);
	const groupWidth = thickness * n + gap * (n - 1);
	const base = l.valuePos(0);
	const out: Column[] = [];
	columnSeries.value.forEach((s, k) => {
		data.value[s].values.forEach((v, c) => {
			if (v === undefined) {
				return;
			}

			const end = l.valuePos(v);
			const x = l.bandStart(c) + (l.bandSize - groupWidth) / 2 + k * (thickness + gap);
			out.push({ series: s, value: v, x, y: Math.min(base, end), w: thickness, h: Math.abs(end - base), end });
		});
	});

	return out;
});

interface Run {
	points: Point[];
	values: number[];
}

// runs are the points of each line and area series, broken where a category has no value.
const runs = computed<Run[][]>(() => {
	const l = layout.value;
	if (!l) {
		return [];
	}

	return data.value.map((d, s) => {
		if (types.value[s] === ChartSeriesTypeValues.ChartSeriesTypeColumn) {
			return [];
		}

		const out: Run[] = [];
		let run: Run | undefined;
		d.values.forEach((v, c) => {
			if (v === undefined) {
				run = undefined;
				return;
			}

			if (!run) {
				run = { points: [], values: [] };
				out.push(run);
			}

			run.points.push({ x: l.bandCenter(c), y: l.valuePos(v) });
			run.values.push(v);
		});

		return out;
	});
});

interface Shape {
	series: number;
	line: string;
	area: string;
	dots: Point[];
}

const shapes = computed<Shape[]>(() => {
	const l = layout.value;
	if (!l) {
		return [];
	}

	const base = l.baseline();
	return runs.value.flatMap((rs, s) => {
		return rs.map((r) => {
			// a line of more points than pixels is drawn through the ones that make a difference
			const points = r.points.length > l.plot.w * 2 ? m4(r.points) : r.points;
			const area = types.value[s] === ChartSeriesTypeValues.ChartSeriesTypeArea;
			return {
				series: s,
				line: points.length > 1 ? linePath(points, props.ui.curve) : '',
				area: area ? areaPath(points, props.ui.curve, base) : '',
				dots: points.length === 1 ? points : [],
			};
		});
	});
});

const markers = computed(() => {
	const l = layout.value;
	const r = markerSize.value;
	if (!l || r <= 0 || l.bandSize < r * 2) {
		return [];
	}

	return runs.value.flatMap((rs, s) => rs.flatMap((run) => run.points.map((p) => ({ series: s, ...p }))));
});

interface ValueLabel extends Label {
	box: [number, number, number, number];
}

// valueLabels writes the values beside the points as long as the series has room for them, leaving out a label
// that would run into one before.
const valueLabels = computed<ValueLabel[]>(() => {
	const l = layout.value;
	if (!l) {
		return [];
	}

	const p = l.plot;
	const out: ValueLabel[] = [];
	const free = (b: ValueLabel['box']) =>
		out.every((o) => b[2] <= o.box[0] || b[0] >= o.box[2] || b[3] <= o.box[1] || b[1] >= o.box[3]);
	const place = (x: number, above: number, below: number, value: number) => {
		const text = formatValue(value, rounding.value);
		const w = measure.value.width(text);
		const half = labelHeight / 2;
		for (const y of [above - half, below + half]) {
			const box: ValueLabel['box'] = [x - w / 2 - 1, y - half, x + w / 2 + 1, y + half];
			if (box[1] >= p.y - 2 && box[3] <= p.y + p.h + 2 && free(box)) {
				out.push({ x, y, text, anchor: 'middle', box });
				return;
			}
		}
	};

	const room = (values: number[]) =>
		Math.max(0, ...values.map((v) => measure.value.width(formatValue(v, rounding.value)))) + 4 <= l.bandSize;

	for (const s of columnSeries.value) {
		const cs = columns.value.filter((c) => c.series === s);
		if (cs.length <= maxLabels && room(cs.map((c) => c.value))) {
			cs.forEach((c) =>
				c.value >= 0
					? place(c.x + c.w / 2, c.end - 1, Infinity, c.value)
					: place(c.x + c.w / 2, -Infinity, c.end + 1, c.value)
			);
		}
	}

	const gap = Math.max(markerSize.value, 2) + 3;
	runs.value.forEach((rs) => {
		const values = rs.flatMap((r) => r.values);
		if (values.length > maxLabels || !room(values)) {
			return;
		}

		rs.forEach((r) => r.points.forEach((pt, i) => place(pt.x, pt.y - gap, pt.y + gap, r.values[i])));
	});

	return out;
});

const inspection = computed(() => {
	const l = layout.value;
	const i = inspected.value;
	if (!l || i < 0 || i >= cats.value.length) {
		return undefined;
	}

	const x = Math.round(l.bandCenter(i)) + 0.5;
	const dots = data.value.flatMap((d, s) => {
		const v = d.values[i];
		if (v === undefined || types.value[s] === ChartSeriesTypeValues.ChartSeriesTypeColumn) {
			return [];
		}

		return [{ series: s, x: l.bandCenter(i), y: l.valuePos(v) }];
	});

	return { x, y1: l.plot.y, y2: l.plot.y + l.plot.h, dots };
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
				kind: kinds.value[s],
			});
		}
	});

	return { title: cats.value[i], rows, x: l.bandCenter(i), top: l.plot.y + 4 };
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
	return named.length > 0 ? `${t('chart.lineChart')}: ${named.join(', ')}` : t('chart.lineChart');
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
			<defs>
				<linearGradient
					v-for="(color, s) in colors"
					:id="`${id}-area-${s}`"
					:key="s"
					gradientUnits="userSpaceOnUse"
					x1="0"
					:y1="layout.plot.y"
					x2="0"
					:y2="layout.baseline()"
				>
					<stop offset="0" :stop-color="color" stop-opacity="0.35" />
					<stop offset="1" :stop-color="color" stop-opacity="0.02" />
				</linearGradient>
			</defs>
			<CartesianAxes :layout="layout" />
			<g>
				<template v-for="(c, i) in columns" :key="i">
					<rect
						v-if="c.h >= 0.5"
						:x="num(c.x)"
						:y="num(c.y)"
						:width="num(c.w)"
						:height="num(c.h)"
						:rx="num(Math.min(3, c.w / 4, c.h / 2))"
						:fill="colors[c.series]"
					/>
				</template>
			</g>
			<g>
				<template v-for="(sh, i) in shapes" :key="i">
					<path v-if="sh.area" :d="sh.area" :fill="`url(#${id}-area-${sh.series})`" />
				</template>
			</g>
			<g fill="none" stroke-width="2" stroke-linejoin="round" stroke-linecap="round">
				<template v-for="(sh, i) in shapes" :key="i">
					<path v-if="sh.line" :d="sh.line" :stroke="colors[sh.series]" />
					<circle
						v-for="(p, k) in sh.dots"
						:key="k"
						:cx="num(p.x)"
						:cy="num(p.y)"
						r="2.5"
						:fill="colors[sh.series]"
						stroke="none"
					/>
				</template>
			</g>
			<g :stroke="markerStroke" stroke-width="2">
				<circle
					v-for="(m, i) in markers"
					:key="i"
					:cx="num(m.x)"
					:cy="num(m.y)"
					:r="markerSize"
					:fill="colors[m.series]"
				/>
			</g>
			<g :font-size="labelSize" fill="currentColor" fill-opacity="0.75">
				<text
					v-for="(v, i) in valueLabels"
					:key="i"
					:x="num(v.x)"
					:y="num(v.y)"
					:text-anchor="v.anchor"
					dominant-baseline="central"
				>
					{{ v.text }}
				</text>
			</g>
			<g v-if="inspection" data-export="false">
				<line
					:x1="inspection.x"
					:y1="inspection.y1"
					:x2="inspection.x"
					:y2="inspection.y2"
					stroke="currentColor"
					stroke-opacity="0.5"
				/>
				<template v-for="d in inspection.dots" :key="d.series">
					<circle :cx="num(d.x)" :cy="num(d.y)" r="5.5" fill="var(--M1)" />
					<circle :cx="num(d.x)" :cy="num(d.y)" r="3.5" :fill="colors[d.series]" />
				</template>
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
