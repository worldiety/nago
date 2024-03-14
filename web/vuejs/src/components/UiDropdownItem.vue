<script setup lang="ts">
import type { LiveDropdownItem } from '@/shared/model';
import { useNetworkStore } from '@/stores/networkStore';

const props = defineProps<{
	ui: LiveDropdownItem;
}>();

const networkStore = useNetworkStore();
</script>

<template>
	<div>
		<p @click="networkStore.invokeFunc(props.ui.onSelected);">{{ props.ui.content.value }}</p>
	</div>
</template>
