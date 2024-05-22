<template>
	<div class="flex flex-wrap gap-4" :class="containerClasses">
		<ui-generic v-for="(element, index) in ui.elements.v" :key="index" :ui="element" :class="elementClass" />
	</div>
</template>

<script setup lang="ts">
import type { FlexContainer } from '@/shared/protocol/ora/flexContainer';
import { computed } from 'vue';
import { Orientation } from '@/shared/protocol/ora/orientation';
import { ContentAlignment } from '@/shared/protocol/ora/contentAlignment';
import UiGeneric from '@/components/UiGeneric.vue';
import { ElementSize } from '@/shared/protocol/ora/elementSize';

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
		default:
			containerClasses.push('justify-start');
			break;
	}
	return containerClasses.join(' ');
});

const elementClass = computed((): string => {
	if (props.ui.orientation.v === Orientation.VERTICAL) {
		return 'w-full';
	}

	switch (props.ui.elementSize.v) {
		case ElementSize.SIZE_SMALL:
			return 'w-64';
		case ElementSize.SIZE_MEDIUM:
			return 'w-80';
		case ElementSize.SIZE_LARGE:
			return 'w-96';
		default:
			return 'w-80';
	}
});
</script>
