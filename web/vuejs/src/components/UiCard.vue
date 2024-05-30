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
		v-if="ui.visible.v"
		class="block rounded-lg p-6 shadow-md bg-primary-94 darkmode:bg-primary-14"
		:class="props.ui.action.v > 0 ? 'cursor-pointer hover:bg-gray-100' : ''"
		@click="onClick"
	>
		<ui-generic v-for="(uiChild, index) in props.ui.children.v" :ui="uiChild" :key="index" />
	</div>
</template>
