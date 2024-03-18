<script setup lang="ts">
import UiDropdownItem from '@/components/UiDropdownItem.vue';
import ArrowDown from '@/assets/svg/arrowDown.svg';
import { computed, onMounted, onUpdated, ref } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveDropdown } from '@/shared/model/liveDropdown';
import type { LiveDropdownItem } from '@/shared/model/liveDropdownItem';

const props = defineProps<{
  ui: LiveDropdown;
}>();

const networkStore = useNetworkStore();
const dropdownOptions = ref<HTMLElement|undefined>();

onMounted(() => {
	if (props.ui.expanded.value) {
		document.addEventListener('click', closeDropdown);
	}
})

onUpdated(() => {
	if (props.ui.expanded.value) {
		document.addEventListener('click', closeDropdown);
	} else {
		document.removeEventListener('click', closeDropdown);
	}
})

const selectedItemNames = computed((): string|null => {
	if (!props.ui.selectedIndexes.value) {
		return null;
	}
	const itemNames = props.ui.items.value
		.filter((item: LiveDropdownItem) => {
		return props.ui.selectedIndexes.value.find((index) => index === item.itemIndex.value) !== undefined;
	}).map((item) => item.content.value);
	return itemNames.length > 0 ? itemNames.join(', ') : null;
});

function closeDropdown(e: MouseEvent) {
	if (e.target instanceof HTMLElement && dropdownOptions.value) {
		const targetHTMLElement = e.target as HTMLElement;
		const dropdownItemWasClicked = targetHTMLElement.compareDocumentPosition(dropdownOptions.value) & Node.DOCUMENT_POSITION_CONTAINS;
		if (!dropdownItemWasClicked) {
			networkStore.invokeFunc(props.ui.onToggleExpanded);
		}
	}
}

function dropdownClicked(forceClose: boolean): void {
	if (!props.ui.disabled.value && (forceClose || !props.ui.expanded.value)) {
		networkStore.invokeFunc(props.ui.onToggleExpanded);
	}
}

function isSelected(itemIndex: number): boolean {
	return props.ui.selectedIndexes.value?.includes(itemIndex) ?? false;
}
</script>

<template>
	<div>
		<span v-if="props.ui.label.value" class="block mb-2 text-sm font-medium">{{ props.ui.label.value }}</span>
		<div class="relative">
			<div
				class="flex justify-between gap-x-4 items-center cursor-default rounded-md p-2"
				:class="props.ui.disabled.value ? 'bg-disabled-background text-disabled-text' : 'border border-black hover:border-wdy-green text-black hover:text-wdy-green'"
				:tabindex="props.ui.disabled.value ? '-1': '0'"
				@click="dropdownClicked(false)"
				@keydown.enter="dropdownClicked(true)"
			>
				<div class="truncate">{{ selectedItemNames ?? 'Auswählen...' }}</div>
				<ArrowDown class="shrink-0 grow-0 duration-100 h-3" :class="{'rotate-180': props.ui.expanded.value}" />
			</div>
			<div ref="dropdownOptions">
				<div v-if="props.ui.expanded.value" class="absolute top-full left-0 right-0 bg-white shadow-lg mt-1 z-40">
					<ui-dropdown-item
						v-for="(dropdownItem, index) in props.ui.items.value"
						:key="index"
						:ui="dropdownItem"
						:multiselect="props.ui.multiselect.value"
						:selected="isSelected(dropdownItem.itemIndex.value)"
					/>
					<div v-if="props.ui.multiselect.value" class="flex justify-center p-2">
						<button class="btn-primary w-full max-w-64" @click="dropdownClicked(true)">Schließen</button>
					</div>
				</div>
			</div>
		</div>
		<!-- Error message has precedence over hints -->
		<p v-if="props.ui.error.value" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>
		<p v-else-if="props.ui.hint.value" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
	</div>
</template>
