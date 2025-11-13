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
import { BarChart, BarChartMarker, ChartDataPoint } from '@/shared/proto/nprotoc_gen';
import { colorToHexValue } from '@/shared/tailwindTranslator';

const props = defineProps<{
	ui: BarChart;
}>();

const options = computed<ApexCharts.ApexOptions>(() => {
	return {
		chart: {
			type: 'bar',
			stacked: props.ui.stacked ?? false,
			toolbar: {
				tools: {
					download: props.ui.chart?.downloadable ?? false,
				},
			},
		},
		plotOptions: {
			bar: {
				horizontal: props.ui.horizontal ?? false,
			},
		},
		colors: colors.value,
		series: series.value,
		noData: {
			text: props.ui.chart?.noDataMessage,
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
		labels: props.ui.chart?.labels?.value ?? [],
	};
});
const colors = computed<string[]>(() => {
	if (!props.ui.chart?.colors) return [];

	return props.ui.chart?.colors.value.map(colorToHexValue).filter((c) => c.length > 0);
});
const series = computed<ApexAxisChartSeries>(() => {
	if (!props.ui.series) return [];

	return props.ui.series.value.map((s, sIndex) => {
		return {
			name: s.label,
			data: s.dataPoints?.value.map((dp, dpIndex) => mapDataPointsToData(dp, sIndex, dpIndex)),
		};
	}) as ApexAxisChartSeries;
});
const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.chart?.frame);

	return styles.join(';');
});

function mapDataPointsToData(dataPoint: ChartDataPoint, seriesIndex: number, dataPointIndex: number) {
	const markers = props.ui.markers?.value.filter(
		(marker) => marker.seriesIndex === seriesIndex && marker.dataPointIndex === dataPointIndex
	);

	return {
		x: dataPoint.x,
		y: dataPoint.y,
		goals: markers?.map(mapMarkerToGoal),
	};
}
function mapMarkerToGoal(marker: BarChartMarker) {
	return {
		name: marker.label,
		value: marker.value,
		strokeDashArray: marker.dashed ? 3 : undefined,
		strokeColor: marker.color ? colorToHexValue(marker.color) : undefined,
		strokeWidth: marker.round ? (props.ui.horizontal ? marker.width : 0) : marker.width,
		strokeHeight: marker.round ? (!props.ui.horizontal ? marker.height : 0) : marker.height,
		strokeLineCap: 'round',
	};
}
</script>

<template>
	<div :style="frameStyles">
		<VueApexCharts type="bar" :series="options.series" :options="options" />
	</div>
</template>
