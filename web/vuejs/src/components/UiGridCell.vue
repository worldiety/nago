<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import { computed } from 'vue';
import {GridCell} from "@/shared/protocol/gen/gridCell";

const props = defineProps<{
	ui: GridCell;
}>();

const styleGridColStart = computed<string>(() => {
	if (props.ui.colStart.v > 0) {
		return `col-start-${props.ui.colStart.v} `;
	}

	return '';
});

const styleGridColEnd = computed<string>(() => {
	if (props.ui.colEnd.v > 0) {
		return `col-end-${props.ui.colEnd.v} `;
	}
	return '';
});

const styleGridRowStart = computed<string>(() => {
	if (props.ui.rowStart.v > 0) {
		return `row-start-${props.ui.rowStart.v} `;
	}

	return '';
});

const styleGridRowEnd = computed<string>(() => {
	if (props.ui.rowEnd.v > 0) {
		return `row-end-${props.ui.rowEnd.v} `;
	}
	return '';
});

const styleGridColSpan = computed<string>(() => {
	let tmp = '';
	if (props.ui.colSpan.v > 0) {
		tmp += `col-span-${props.ui.colSpan.v} `;
	}

	if (props.ui.smColSpan.v > 0) {
		tmp += ` sm:col-span-${props.ui.smColSpan.v} `;
	}

	if (props.ui.mdColSpan.v > 0) {
		tmp += ` md:col-span-${props.ui.mdColSpan.v} `;
	}

	if (props.ui.lgColSpan.v > 0) {
		tmp += ` lg:col-span-${props.ui.lgColSpan.v} `;
	}

	return tmp;
});

const style = computed(
	() => `
    ${styleGridColStart.value}
    ${styleGridColEnd.value}
    ${styleGridRowStart.value}
    ${styleGridRowEnd.value}
    ${styleGridColSpan.value}
`
);
</script>

<template>
	<div :class="style">
		<ui-generic :ui="props.ui.body.v" />
	</div>
</template>
