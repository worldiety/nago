<script setup lang="ts">
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveDropdownItem } from '@/shared/model/liveDropdownItem';

const props = defineProps<{
	ui: LiveDropdownItem;
}>();

const networkStore = useNetworkStore();
</script>

<template>
	<div class="bg-white">
		<div
			class="cursor-default p-2 hover:text-wdy-green hover:bg-wdy-green hover:bg-opacity-5"
			@click="networkStore.invokeFunc(props.ui.onSelected);"
		>
			<p class="truncate pl-2">{{ props.ui.content.value }}</p>
		</div>
	</div>
</template>

<style scoped>

</style>
