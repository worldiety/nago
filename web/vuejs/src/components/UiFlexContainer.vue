<template>
	<div v-if="ui.visible.v" class="flex flex-wrap gap-4" :class="containerClasses">
		<ui-generic v-for="(element, index) in ui.elements.v" :key="index" :ui="element" :class="elementClasses" />
	</div>
</template>

<script setup lang="ts">
import type { FlexContainer } from '@/shared/protocol/ora/flexContainer';
import { computed } from 'vue';
import { Orientation } from '@/shared/protocol/ora/orientation';
import { ContentAlignment } from '@/shared/protocol/ora/contentAlignment';
import UiGeneric from '@/components/UiGeneric.vue';
import { ElementSize } from '@/shared/protocol/ora/elementSize';
import { ItemsAlignment } from '@/shared/protocol/ora/itemsAlignment';

const props = defineProps<{
	ui: FlexContainer;
}>();

const containerClasses = computed((): string => {
	const containerClasses: string[] = [];
	switch (props.ui.orientation.v) {
		case Orientation.HORIZONTAL:
			containerClasses.push('flex-row');
			break;
		case Orientation.VERTICAL:
			containerClasses.push('flex-col');
			break;
		default:
			containerClasses.push('flex-row');
			break;
	}

	switch (props.ui.contentAlignment.v) {
		case ContentAlignment.START:
			containerClasses.push('justify-start');
			break;
		case ContentAlignment.CENTER:
			containerClasses.push('justify-center');
			break;
		case ContentAlignment.END:
			containerClasses.push('justify-end');
			break;
		case ContentAlignment.BETWEEN:
			containerClasses.push('justify-between');
			break;
		default:
			containerClasses.push('justify-center');
			break;
	}

	switch (props.ui.itemsAlignment.v) {
		case ItemsAlignment.START:
			containerClasses.push('items-start');
			break;
		case ItemsAlignment.CENTER:
			containerClasses.push('items-center');
			break;
		case ItemsAlignment.END:
			containerClasses.push('items-end');
			break;
		case ItemsAlignment.STRETCH:
			containerClasses.push('items-stretch');
			break;
		default:
			containerClasses.push('items-stretch');
			break;
	}

	return containerClasses.join(' ');
});

const elementClasses = computed((): string => {
	const elementClasses: string[] = [];
	if (props.ui.orientation.v === Orientation.VERTICAL) {
		elementClasses.push('w-full');
	}

	switch (props.ui.elementSize.v) {
		case ElementSize.SIZE_AUTO:
			elementClasses.push('basis-auto');
			break;
		case ElementSize.SIZE_SMALL:
			elementClasses.push('basis-64');
			break;
		case ElementSize.SIZE_MEDIUM:
			elementClasses.push('basis-80');
			break;
		case ElementSize.SIZE_LARGE:
			elementClasses.push('basis-96');
			break;
		default:
			elementClasses.push('basis-auto');
			break;
	}
	return elementClasses.join(' ');
});
</script>
