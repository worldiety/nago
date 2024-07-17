<script lang="ts" setup>
import {computed} from 'vue';
import UiGridCell from '@/components/UiGridCell.vue';
import {Grid} from "@/shared/protocol/ora/grid";

const props = defineProps<{
	ui: Grid;
}>();


const style = computed<string>(() => {
	const styles: string[] = [];

	if (!props.ui.c || props.ui.c === 0) {
		styles.push("grid-auto-columns: auto")
	} else {
		styles.push(`grid-template-columns: repeat(${props.ui.c}, minmax(0, 1fr))`)
	}

	if (!props.ui.r || props.ui.r === 0) {
		styles.push("grid-auto-rows: auto")
	} else {
		styles.push(`grid-template-rows: repeat(${props.ui.r}, minmax(0, 1fr))`)
	}

	if (props.ui.g) {
		styles.push(`gap: ${props.ui.g}`)
	}

	return styles.join(";");
});
</script>

<template>
	<div class="grid" :style="style">
		<ui-grid-cell v-for="cell in props.ui.b" :ui="cell"/>
	</div>
</template>
