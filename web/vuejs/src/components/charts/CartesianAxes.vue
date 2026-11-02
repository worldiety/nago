<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import type { Cartesian } from '@/components/charts/cartesian';
import { labelSize, num, titleSize } from '@/components/charts/chart';
import ChartLegend from '@/components/charts/ChartLegend.vue';

const props = defineProps<{
	layout: Cartesian;
}>();

// crisp moves a hairline onto the middle of a pixel.
function crisp(v: number): number {
	return Math.round(v) + 0.5;
}

// line returns the hairline across the plot at the position on the value axis.
function line(pos: number) {
	const p = props.layout.plot;
	const v = crisp(pos);
	return props.layout.horizontal
		? { x1: v, y1: num(p.y), x2: v, y2: num(p.y + p.h) }
		: { x1: num(p.x), y1: v, x2: num(p.x + p.w), y2: v };
}

const grid = computed(() => props.layout.ticks.map((t) => line(t.pos)));
const baseline = computed(() => line(props.layout.baseline()));
</script>

<template>
	<g>
		<g stroke="currentColor" stroke-opacity="0.12">
			<line v-for="(l, i) in grid" :key="i" v-bind="l" />
		</g>
		<line v-bind="baseline" stroke="currentColor" stroke-opacity="0.35" />
		<g :font-size="labelSize" fill="currentColor" fill-opacity="0.65">
			<text
				v-for="(l, i) in layout.tickLabels"
				:key="`t${i}`"
				:x="num(l.x)"
				:y="num(l.y)"
				:text-anchor="l.anchor"
				dominant-baseline="central"
				>{{ l.text }}</text
			>
			<text
				v-for="(l, i) in layout.categoryLabels"
				:key="`c${i}`"
				:x="num(l.x)"
				:y="num(l.y)"
				:text-anchor="l.anchor"
				dominant-baseline="central"
				>{{ l.text }}</text
			>
		</g>
		<g :font-size="titleSize" font-weight="500" fill="currentColor" fill-opacity="0.65">
			<text
				v-for="(l, i) in layout.titles"
				:key="i"
				:x="num(l.x)"
				:y="num(l.y)"
				:text-anchor="l.anchor"
				dominant-baseline="central"
				>{{ l.text }}</text
			>
		</g>
		<ChartLegend :items="layout.legend" />
	</g>
</template>
