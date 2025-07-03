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
import { BarChart, BarChartDataPoint, BarChartMarker } from '@/shared/proto/nprotoc_gen';
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
					download: props.ui.downloadable ?? false,
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
			text: props.ui.noDataMessage,
		},
		labels: props.ui.labels?.value ?? [],
	};
});
const colors = computed<string[]>(() => {
	if (!props.ui.colors) return [];

	return props.ui.colors.value.map(colorToHexValue).filter((c) => c.length > 0);
});
const series = computed<ApexAxisChartSeries>(() => {
	if (!props.ui.series) return [];

	return props.ui.series.value.map((s) => {
		return {
			name: s.label,
			data: s.dataPoints?.value.map(mapDataPointsToData),
		};
	}) as ApexAxisChartSeries;
});
const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);

	return styles.join(';');
});

function mapDataPointsToData(dataPoint: BarChartDataPoint) {
	return {
		x: dataPoint.x,
		y: dataPoint.y,
		goals: dataPoint.markers ? dataPoint.markers.value.map(mapMarkerToGoal) : undefined,
	};
}
function mapMarkerToGoal(marker: BarChartMarker) {
	return {
		name: marker.label,
		value: marker.value,
		strokeDashArray: marker.isDashed ? 3 : undefined,
		strokeColor: marker.color ? colorToHexValue(marker.color) : undefined,
		strokeWidth: marker.isRound ? (props.ui.horizontal ? marker.width : 0) : marker.width,
		strokeHeight: marker.isRound ? (!props.ui.horizontal ? marker.height : 0) : marker.height,
		strokeLineCap: 'round',
	};
}
</script>

<template>
	<div :style="frameStyles">
		<VueApexCharts type="bar" :series="options.series" :options="options" />
	</div>
</template>
