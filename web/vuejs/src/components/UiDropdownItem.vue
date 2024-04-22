<script setup lang="ts">
import { useNetworkStore } from '@/stores/networkStore';
import type {DropdownItem} from "@/shared/protocol/gen/dropdownItem";

const props = defineProps<{
	ui: DropdownItem;
	multiselect: boolean;
	selected: boolean;
}>();

const networkStore = useNetworkStore();
</script>

<template>
	<div
		class="cursor-default mx-2.5 py-4
					hover:text-ora-orange hover:bg-ora-orange hover:rounded-2lg hover:bg-opacity-15
					dark:hover:bg-ora-orange dark:hover:rounded-2lg dark:text-white dark:hover:text-ora-orange dark:hover:bg-opacity-25"
		tabindex="0"
		@click="networkStore.invokeFunctions(props.ui.onClicked)"
		@keydown.enter="networkStore.invokeFunctions(props.ui.onClicked)"
	>
			<div class="flex justify-start items-center pl-2.5">
				<input v-if="props.multiselect" type="checkbox" tabindex="-1" :checked="props.selected" class="focus:ring-0">
				<div v-if="props.multiselect">
					<p class="truncate pl-2">{{ props.ui.content.v }}</p>
				</div>
				<div v-else>
					<p>{{ props.ui.content.v }}</p>
				</div>
			</div>
	</div>
</template>

<style scoped></style>
