<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import { Chart } from '@/components/charts/chart';
import { frameCSS } from '@/components/shared/frame';
import { cssLengthValue } from '@/components/shared/length';
import type ApexCharts from 'apexcharts';
import VueApexCharts from 'vue3-apexcharts';
import type { ChartDataPoint, LineChart } from '@/shared/proto/nprotoc_gen';
import { ChartSeriesTypeValues, LineChartCurveValues } from '@/shared/proto/nprotoc_gen';
import { colorToHexValue } from '@/shared/tailwindTranslator';
import { ThemeKey, useThemeManager } from '@/shared/themeManager';

const props = defineProps<{
	ui: LineChart;
}>();

const themeManager = useThemeManager();

const options = computed<ApexCharts.ApexOptions>(() => {
	return {
		chart: {
			type: 'line',
			foreColor: 'currentColor',
			toolbar: {
				tools: {
					download: props.ui.chart?.downloadable ?? false,
				},
			},
			height: cssLengthValue(props.ui.chart?.frame?.height ?? 'auto'),
			width: cssLengthValue(props.ui.chart?.frame?.width ?? 'auto'),
		},
		tooltip: {
			theme: themeManager.getActiveThemeKey() === ThemeKey.DARK ? 'dark' : 'light',
		},
		colors: colors.value,
		series: series.value,
		noData: {
			text: props.ui.chart?.noDataMessage,
		},
		labels: props.ui.chart?.labels?.value ?? [],
		stroke: {
			curve: mapCurve(props.ui.curve),
		},
		xaxis: {
			title: {
				text: props.ui.chart?.xAxisTitle,
			},
		},
		yaxis: {
			title: {
				text: props.ui.chart?.yAxisTitle,
			},
		},
		markers: {
			size: props.ui.markers?.size,
			strokeColors: props.ui.markers?.borderColor ? colorToHexValue(props.ui.markers?.borderColor) : '#fff',
			showNullDataPoints: props.ui.markers?.showNullDataPoints,
		},
		dataLabels: {
			formatter: Chart.DataLabelFormatter(props.ui.chart),
		},
	};
});
const colors = computed<string[]>(() => {
	if (!props.ui.chart?.colors) return [];

	return props.ui.chart.colors.value.map(colorToHexValue).filter((c) => c.length > 0);
});
const series = computed<ApexAxisChartSeries>(() => {
	if (!props.ui.series) return [];

	return props.ui.series.value.map((s) => {
		return {
			name: s.label,
			type: mapSeriesType(s.type),
			data: s.dataPoints?.value.map(mapDataPointsToData),
		};
	}) as ApexAxisChartSeries;
});
const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.chart?.frame);

	return styles.join(';');
});

function mapDataPointsToData(dataPoint: ChartDataPoint) {
	return {
		x: dataPoint.x,
		y: dataPoint.y,
	};
}

function mapCurve(curve: number | undefined) {
	switch (curve) {
		case LineChartCurveValues.LineChartCurveStraight:
			return 'straight';
		case LineChartCurveValues.LineChartCurveSmooth:
			return 'smooth';
		case LineChartCurveValues.LineChartCurveStepline:
			return 'stepline';
		default:
			return 'straight';
	}
}

function mapSeriesType(seriesType: number | undefined) {
	switch (seriesType) {
		case ChartSeriesTypeValues.ChartSeriesTypeLine:
			return 'line';
		case ChartSeriesTypeValues.ChartSeriesTypeColumn:
			return 'column';
		case ChartSeriesTypeValues.ChartSeriesTypeArea:
			return 'area';
		default:
			return 'line';
	}
}
</script>

<template>
	<div :style="frameStyles">
		<VueApexCharts type="line" :series="options.series" :options="options" />
	</div>
</template>
