<script lang="ts" setup>
import { computed } from 'vue';
import UiGridCell from '@/components/UiGridCell.vue';
import { gapSize2Tailwind } from '@/shared/tailwindTranslator';
import type { LivePage } from '@/shared/model/livePage';
import {Grid} from "@/shared/protocol/gen/grid";

const props = defineProps<{
	ui: Grid;
	page: LivePage;
}>();

//TODO we get into trouble using tailwind pre-processor here
const style = computed<string>(() => {
	let tmp = 'grid';
	if (props.ui.columns.v > 0) {
		tmp += ` grid-cols-${props.ui.columns.v}`;
	} else {
		if (props.ui.rows.v > 0) {
			tmp += ' grid-flow-col';
		} else {
			tmp += ' grid-cols-auto';
		}
	}

	if (props.ui.smColumns.v > 0) {
		tmp += ` sm:grid-cols-${props.ui.smColumns.v}`;
	}

	if (props.ui.mdColumns.v > 0) {
		tmp += ` md:grid-cols-${props.ui.mdColumns.v}`;
	}

	if (props.ui.lgColumns.v > 0) {
		tmp += ` lg:grid-cols-${props.ui.lgColumns.v}`;
	}

	if (props.ui.rows.v > 0) {
		tmp += ` grid-rows-${props.ui.rows.v}`;
	} else {
		tmp += ' grid-rows-auto';
	}

	tmp += ' ' + gapSize2Tailwind(props.ui.gap.v);

	return tmp;
});
</script>

<template>
	<div :class="style">
		<ui-grid-cell v-for="cell in props.ui.cells.v" :ui="cell" :page="page" />
	</div>
</template>
