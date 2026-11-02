<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { type PlacedLegendEntry, labelHeight, labelSize, legendSwatch, num } from '@/components/charts/chart';

defineProps<{
	items: PlacedLegendEntry[];
}>();
</script>

<template>
	<g :font-size="labelSize">
		<g v-for="(e, i) in items" :key="i" :transform="`translate(${num(e.x)} ${num(e.y + labelHeight / 2)})`">
			<circle v-if="e.kind === 'dot'" cx="5" cy="0" r="4" :fill="e.color" />
			<rect v-else x="1" y="-4" width="8" height="8" rx="2" :fill="e.color" />
			<text :x="legendSwatch" y="0" dominant-baseline="central" fill="currentColor" fill-opacity="0.65">{{
				e.label
			}}</text>
		</g>
	</g>
</template>
