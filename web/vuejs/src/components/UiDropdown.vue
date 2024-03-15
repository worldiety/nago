<script setup lang="ts">
import UiDropdownItem from '@/components/UiDropdownItem.vue';
import ArrowDown from '@/assets/svg/arrowDown.svg';
import { computed } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveDropdown } from '@/shared/model/liveDropdown';
import type { LiveDropdownItem } from '@/shared/model/liveDropdownItem';

const props = defineProps<{
  ui: LiveDropdown;
}>();

const networkStore = useNetworkStore();

const selectedItemName = computed((): string => {
	return props.ui.items.value.find((item: LiveDropdownItem) => item.itemIndex.value === props.ui.selectedIndex.value)?.content.value ?? '';
});

function dropdownClicked(): void {
	if (!props.ui.disabled.value) {
		networkStore.invokeFunc(props.ui.onToggleExpanded);
	}
}
</script>

<template>
	<div>
		<span v-if="props.ui.label.value" class="block mb-2 text-sm font-medium">{{ props.ui.label.value }}</span>
		<div class="flex flex-col gap-y-1">
			<div
				class="flex justify-between gap-x-4 items-center cursor-default rounded-md p-2"
				:class="props.ui.disabled.value ? 'bg-disabled-background text-disabled-text' : 'border border-black hover:border-wdy-green text-black hover:text-wdy-green'"
				@click="dropdownClicked"
			>
				<div class="truncate">{{ selectedItemName }}</div>
				<ArrowDown class="duration-100 h-3" :class="{'rotate-180': props.ui.expanded.value}" />
			</div>
			<div v-if="props.ui.expanded.value" class="shadow-lg">
				<ui-dropdown-item
					v-for="(dropdownItem, index) in props.ui.items.value"
					:key="index"
					:ui="dropdownItem"
				/>
			</div>
		</div>
		<!-- Error message has precedence over hints -->
		<p v-if="props.ui.error.value" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>
		<p v-else-if="props.ui.hint.value" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
	</div>
</template>
