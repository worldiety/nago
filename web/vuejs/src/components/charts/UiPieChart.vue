<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { Chart } from '@/components/charts/chart';
import { frameCSS } from '@/components/shared/frame';
import { randomStr } from '@/components/shared/util';
import type ApexCharts from 'apexcharts';
import VueApexCharts from 'vue3-apexcharts';
import type { PieChart } from '@/shared/proto/nprotoc_gen';
import { colorToHexValue } from '@/shared/tailwindTranslator';
import { ThemeKey, useThemeManager } from '@/shared/themeManager';

const props = defineProps<{
	ui: PieChart;
}>();

const themeManager = useThemeManager();

const id = randomStr(16);
const refreshCount = ref(0);
const refreshKey = computed<string>(() => `${id}_${refreshCount.value}`);

const chartType = computed<string>(() => {
	return props.ui.showAsDonut ? 'donut' : 'pie';
});

const options = computed<ApexCharts.ApexOptions>(() => {
	return {
		chart: {
			foreColor: 'currentColor',
			toolbar: {
				tools: {
					download: props.ui.chart?.downloadable ?? false,
				},
			},
		},
		tooltip: {
			theme: themeManager.getActiveThemeKey() === ThemeKey.DARK ? 'dark' : 'light',
		},
		colors: colors.value,
		noData: {
			text: props.ui.chart?.noDataMessage,
		},
		labels: props.ui.chart?.labels?.value ?? labelsFromData.value ?? [],
		dataLabels: {
			formatter(val: string | number | number[], { seriesIndex, w }): string | number | (string | number)[] {
				if (props.ui.showAbsoluteValues) {
					if (typeof val !== 'number') return w.config.series[seriesIndex];
					val = w.config.series[seriesIndex];
				}

				const formatter = Chart.DataLabelFormatter(props.ui.chart);
				return props.ui.showDataLabels ? formatter(val) : '';
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

		return s.dataPoints.value.map((dp) => dp.y ?? 0);
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

function refreshChart() {
	refreshCount.value++;
}

watch(() => props.ui, refreshChart, { deep: true });
</script>

<template>
	<div :style="frameStyles">
		<VueApexCharts :key="refreshKey" :type="chartType" :series="series" :options="options" />
	</div>
</template>
