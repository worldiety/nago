<script lang="ts" setup>
import type { Ref } from 'vue';
import { inject } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { useNetworkStore } from '@/stores/networkStore';
import {Card} from "@/shared/protocol/gen/card";

const props = defineProps<{
	ui: Card;
}>();

const networkStore = useNetworkStore();

function onClick() {
	networkStore.invokeFunctions(props.ui.action);
}
</script>

<template>
	<div
		@click="onClick"
		:class="props.ui.action.v > 0 ? 'cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700' : ''"
		class="block max-w-sm rounded-lg border border-gray-200 bg-white p-6 shadow dark:border-gray-700 dark:bg-gray-800"
	>
		<ui-generic v-for="ui in props.ui.children.v" :ui="ui" />
	</div>
</template>
