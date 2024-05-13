<template>
	<div
		class="flex justify-start items-center gap-x-2 hover:text-ora-orange focus-visible:text-ora-orange hover:bg-ora-orange hover:bg-opacity-15 focus-visible:bg-ora-orange focus-visible:bg-opacity-15 active:bg-opacity-25 cursor-pointer rounded-full overflow-hidden px-3"
		tabindex="0"
		@click="executeAction"
		@keydown.enter="executeAction"
	>
		<div v-if="icon" class="h-4 *:h-full" v-html="icon"></div>
		<span class="select-none truncate">{{ breadcrumbItem.label.v }}</span>
	</div>
</template>

<script setup lang="ts">
import type { BreadcrumbItem } from '@/shared/protocol/gen/breadcrumbItem';
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	breadcrumbItem: BreadcrumbItem;
	icon: string|null;
}>();

const serviceAdapter = useServiceAdapter();

function executeAction(): void {
	serviceAdapter.executeFunctions(props.breadcrumbItem.action);
}
</script>
