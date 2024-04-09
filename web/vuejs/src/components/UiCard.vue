<script lang="ts" setup>
import type { Ref } from 'vue';
import { inject } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveCard } from '@/shared/model/liveCard';
import type { LivePage } from '@/shared/model/livePage';
import type { UiDescription } from '@/shared/model/uiDescription';

const props = defineProps<{
	ui: LiveCard;
	page: LivePage;
}>();

const networkStore = useNetworkStore();
const ui: Ref<UiDescription> = inject('ui')!;

function onClick() {
	networkStore.invokeFunctions(props.ui.action);
}
</script>

<template>
	<div
		@click="onClick"
		:class="props.ui.action.value > 0 ? 'cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700' : ''"
		class="block max-w-sm rounded-lg border border-gray-200 bg-white p-6 shadow dark:border-gray-700 dark:bg-gray-800"
	>
		<ui-generic v-for="ui in props.ui.children.value" :ui="ui" :page="page" />
	</div>
</template>
