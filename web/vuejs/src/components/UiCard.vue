<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import type {Card} from "@/shared/protocol/ora/card";
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: Card;
}>();

const serviceAdapter = useServiceAdapter();

function onClick() {
	serviceAdapter.executeFunctions(props.ui.action);
}
</script>

<template>
	<div
		@click="onClick"
		:class="props.ui.action.v > 0 ? 'cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700' : ''"
		class="block rounded-lg border border-gray-200 bg-white p-6 shadow dark:border-gray-700 dark:bg-gray-800"
	>
		<ui-generic v-for="(uiChild, index) in props.ui.children.v" :ui="uiChild" :key="index" />
	</div>
</template>
