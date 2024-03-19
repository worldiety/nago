<script setup lang="ts">
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveDropdownItem } from '@/shared/model/liveDropdownItem';

const props = defineProps<{
	ui: LiveDropdownItem;
	multiselect: boolean;
	selected: boolean;
}>();

const networkStore = useNetworkStore();
</script>

<template>
	<div
		class="cursor-default p-2 hover:text-wdy-green hover:bg-wdy-green hover:bg-opacity-5"
		tabindex="0"
		@click="networkStore.invokeFunc(props.ui.onClicked)"
		@keydown.enter="networkStore.invokeFunc(props.ui.onClicked)"
	>
		<div class="flex justify-start items-center">
			<input v-if="props.multiselect" type="checkbox" tabindex="-1" :checked="props.selected" class="focus:ring-0">
			<p class="truncate pl-2">{{ props.ui.content.value }}</p>
		</div>
	</div>
</template>

<style scoped></style>
