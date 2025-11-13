<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import { frameCSS } from '@/components/shared/frame';
import type ApexCharts from 'apexcharts';
import VueApexCharts from 'vue3-apexcharts';
import { PieChart } from '@/shared/proto/nprotoc_gen';
import { colorToHexValue } from '@/shared/tailwindTranslator';

const props = defineProps<{
	ui: PieChart;
}>();

const chartType = computed<string>(() => {
	return props.ui.showAsDonut ? 'donut' : 'pie';
});
const options = computed<ApexCharts.ApexOptions>(() => {
	return {
		chart: {
			toolbar: {
				tools: {
					download: props.ui.chart?.downloadable ?? false,
				},
			},
		},
		colors: colors.value,
		series: series.value,
		noData: {
			text: props.ui.chart?.noDataMessage,
		},
		labels: props.ui.chart?.labels?.value ?? labelsFromData.value ?? [],
		dataLabels: {
			formatter(val: string | number | number[]): string | number | (string | number)[] {
				return props.ui.showDataLabels ? val.toString() : '';
			},
		},
	};
});
const colors = computed<string[]>(() => {
	if (!props.ui.chart?.colors) return [];

	return props.ui.chart?.colors.value.map(colorToHexValue).filter((c) => c.length > 0);
});
const series = computed<ApexNonAxisChartSeries>(() => {
	if (!props.ui.series || props.ui.series.value.length === 0) return [];

	const apexAxisChartSeriesRaw = props.ui.series.value.map((s) => {
		if (!s.dataPoints || s.dataPoints.value.length === 0) return [];

		return s.dataPoints.value.map((dp) => dp.y ?? null);
	});

	if (apexAxisChartSeriesRaw.length > 1) {
		console.warn('Multiple series are not supported. Only the first series will be used for data points.');
	}

	return apexAxisChartSeriesRaw[0] as ApexNonAxisChartSeries;
});
const labelsFromData = computed<string[]>(() => {
	if (!props.ui.series || props.ui.series.value.length === 0) return [];

	const apexAxisChartSeriesRaw = props.ui.series.value.map((s) => {
		if (!s.dataPoints || s.dataPoints.value.length === 0) return [];

		return s.dataPoints.value.map((dp) => dp.x ?? '');
	});

	if (apexAxisChartSeriesRaw.length > 1) {
		console.warn('Multiple series are not supported. Only the first series will be used for labels.');
	}

	return apexAxisChartSeriesRaw[0];
});
const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.chart?.frame);

	return styles.join(';');
});
</script>

<template>
	<div :style="frameStyles">
		<VueApexCharts :type="chartType" :series="options.series" :options="options" />
	</div>
</template>
