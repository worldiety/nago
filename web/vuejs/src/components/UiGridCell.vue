<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import { computed } from 'vue';
import {GridCell} from "@/shared/protocol/ora/gridCell";

const props = defineProps<{
	ui: GridCell;
}>();

const style = computed<string>(() => {
	const styles: string[] = [];

	if (props.ui.cs) {
		styles.push(`grid-column-start: ${props.ui.cs}`)
	}

	if (props.ui.ce) {
		styles.push(`grid-column-end: ${props.ui.ce}`)
	}

	if (props.ui.rs) {
		styles.push(`grid-row-start: ${props.ui.rs}`)
	}

	if (props.ui.re) {
		styles.push(`grid-row-end: ${props.ui.re}`)
	}

	if (props.ui.cp) {
		styles.push(`grid-column: span ${props.ui.cp} / span ${props.ui.cp}`)
	}

	if (props.ui.rp) {
		styles.push(`grid-row: span ${props.ui.cp} / span ${props.ui.cp}`)
	}


	return styles.join(";");
});
</script>

<template>
		<ui-generic v-if="props.ui.b" :ui="props.ui.b" :style="style"/>
</template>
