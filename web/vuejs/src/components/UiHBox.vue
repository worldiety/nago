<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import { computed } from 'vue';
import {HBox} from "@/shared/protocol/ora/hBox";

const props = defineProps<{
	ui: HBox;
}>();

const clazz = computed<string>(() => {
	switch (props.ui.alignment.v) {
		case 'flex-center':
			return 'flex justify-center';
		case 'flex-left':
			return 'flex';
		case 'flex-right':
			return 'flex flex-row-reverse';
		case 'grid':
		default:
			return 'grid grid-flow-col grid-cols-auto gap-x-3';
	}
});
</script>

<template>
	<div :class="clazz">
		<ui-generic v-for="ui in props.ui.children.v" :ui="ui"  />
	</div>
</template>
