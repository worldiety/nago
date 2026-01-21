<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { colorValue, opacityValue } from '@/components/shared/colors';
import { marginCSS } from '@/components/shared/padding';
import { AlignmentValues, GridCell } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: GridCell;
}>();

const style = computed<string>(() => {
	const styles: string[] = [];

	if (props.ui.rowStart) {
		styles.push(`grid-row-start: ${props.ui.rowStart}`);
	}

	if (props.ui.rowEnd) {
		styles.push(`grid-row-end: ${props.ui.rowEnd}`);
	}

	if (props.ui.colSpan) {
		styles.push(`grid-column: span ${props.ui.colSpan} / span ${props.ui.colSpan}`);
	}

	if (props.ui.rowSpan) {
		styles.push(`grid-row: span ${props.ui.rowSpan} / span ${props.ui.rowSpan}`);
	}

	if (props.ui.colStart) {
		styles.push(`grid-column-start: ${props.ui.colStart}`);
	}

	if (props.ui.colEnd) {
		styles.push(`grid-column-end: ${props.ui.colEnd}`);
	}

	if (props.ui.backgroundColor) {
		const color = colorValue(props.ui.backgroundColor);
		styles.push(`background-color: ${color}`);
	}

	if (props.ui.alignment === undefined) {
		props.ui.alignment = AlignmentValues.Center;
	}

	switch (props.ui.alignment) {
		case AlignmentValues.Stretch:
			// do nothing, which stretches
			break;
		case AlignmentValues.Center:
			styles.push('place-self: center');
			break;
		case AlignmentValues.Leading:
			styles.push('place-self: center start');
			break;
		case AlignmentValues.Trailing:
			styles.push('place-self: center end');
			break;
		case AlignmentValues.Top:
			styles.push('place-self: start center');
			break;
		case AlignmentValues.Bottom:
			styles.push('place-self: end center');
			break;
		case AlignmentValues.TopLeading:
			styles.push('place-self: start');
			break;
		case AlignmentValues.TopTrailing:
			styles.push('place-self: start end');
			break;
		case AlignmentValues.BottomLeading:
			styles.push('place-self: end start');
			break;
		case AlignmentValues.BottomTrailing:
			styles.push('place-self: end end');
			break;
	}

	styles.push(...marginCSS(props.ui.padding));

	return styles.join(';');
});
</script>

<template>
	<!-- gridcell -->
	<ui-generic v-if="props.ui.body" :ui="props.ui.body" :style="style" />
</template>
