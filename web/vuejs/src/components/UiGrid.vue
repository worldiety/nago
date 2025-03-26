<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import UiGridCell from '@/components/UiGridCell.vue';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { fontCSS } from '@/components/shared/font';
import { frameCSS } from '@/components/shared/frame';
import { cssLengthValue } from '@/components/shared/length';
import { paddingCSS } from '@/components/shared/padding';
import { Grid } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Grid;
}>();

const style = computed<string>(() => {
	const styles: string[] = [];

	if (!props.ui.columns) {
		styles.push('grid-auto-columns: auto');
	} else {
		if (!props.ui.colWidths) {
			styles.push(`grid-template-columns: repeat(${props.ui.columns}, minmax(0, 1fr))`);
		} else {
			let tmp = 'grid-template-columns: ';
			for (let len of props.ui.colWidths.value) {
				tmp += cssLengthValue(len);
				tmp += ' ';
			}

			let restColCount = props.ui.columns - props.ui.colWidths.value.length;
			if (restColCount > 0) {
				tmp += `repeat(${restColCount}, 1fr)`;
			}

			styles.push(tmp);
		}
	}

	if (!props.ui.rows) {
		styles.push('grid-auto-rows: auto');
	} else {
		styles.push(`grid-template-rows: repeat(${props.ui.rows}, minmax(0, 1fr))`);
	}

	if (props.ui.rowGap) {
		styles.push(`row-gap: ${props.ui.rowGap}`);
	}

	if (props.ui.colGap) {
		styles.push(`column-gap: ${props.ui.colGap}`);
	}

	if (props.ui.backgroundColor) {
		styles.push(`background-color: ${colorValue(props.ui.backgroundColor)}`);
	}

	styles.push(...frameCSS(props.ui.frame));
	styles.push(...borderCSS(props.ui.border));
	styles.push(...paddingCSS(props.ui.padding));
	styles.push(...fontCSS(props.ui.font));

	return styles.join(';');
});

const clazz = computed<string>(() => {
	const styles: string[] = [];
	styles.push('grid overflow-clip');

	return styles.join(' ');
});
</script>

<template>
	<!-- grid -->
	<div :class="clazz" :style="style">
		<ui-grid-cell v-for="cell in props.ui.cells?.value" :ui="cell" />
	</div>
</template>
