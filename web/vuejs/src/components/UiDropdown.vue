<script setup lang="ts">
import UiDropdownItem from '@/components/UiDropdownItem.vue';
import ArrowDown from '@/assets/svg/arrowDown.svg';
import { computed, onMounted, onUpdated, ref } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import {Dropdown} from "@/shared/protocol/gen/dropdown";
import {DropdownItem} from "@/shared/protocol/gen/dropdownItem";

const props = defineProps<{
	ui: Dropdown;
}>();

const networkStore = useNetworkStore();
const dropdownOptions = ref<HTMLElement|undefined>();

onMounted(() => {
	if (props.ui.expanded.v) {
		document.addEventListener('click', closeDropdown);
	}
})

onUpdated(() => {
	document.removeEventListener('click', closeDropdown);
	if (props.ui.expanded.v) {
		document.addEventListener('click', closeDropdown);
	}
})

const selectedItemNames = computed((): string|null => {
	if (!props.ui.selectedIndices.v) {
		return null;
	}
	const itemNames = props.ui.items.v
		.filter((item: DropdownItem) => {
			const itemIndex = indexOf(item);
			return props.ui.selectedIndices.v.find((index) => index === itemIndex) !== undefined;
	}).map((item) => item.content.v);
	return itemNames.length > 0 ? itemNames.join(', ') : null;
});

/**
 * Determines the index of a dropdown item based on its ID
 * 
 * @param item The dropdown item to determine the index of
 */
function indexOf(item: DropdownItem): number {
	return props.ui.items.v.findIndex((it) => it.id == item.id) ?? -1;
}

function closeDropdown(e: MouseEvent) {
	e.preventDefault();
	if (e.target instanceof HTMLElement && dropdownOptions.value) {
		const targetHTMLElement = e.target as HTMLElement;
		const dropdownItemWasClicked = targetHTMLElement.compareDocumentPosition(dropdownOptions.value) & Node.DOCUMENT_POSITION_CONTAINS;
		if (!dropdownItemWasClicked) {
			networkStore.invokeFunctions(props.ui.onClicked);
		}
	}
}

function dropdownClicked(forceClose: boolean): void {
	if (!props.ui.disabled.v && (forceClose || !props.ui.expanded.v)) {
		networkStore.invokeFunctions(props.ui.onClicked);
	}
}

function isSelected(item: DropdownItem): boolean {
	const itemIndex = indexOf(item);
	return props.ui.selectedIndices.v?.includes(itemIndex) ?? false;
}
</script>

<template>
	<div>
		<div class="relative">
			<!-- Input field -->
			<InputWrapper
				:label="props.ui.label.v"
				:error="props.ui.error.v"
				:hint="props.ui.hint.v"
				:disabled="props.ui.disabled.v"
			>
				<div
					class="input-field flex justify-between gap-x-4 items-center cursor-default"
					:tabindex="props.ui.disabled.v ? '-1': '0'"
					@click="dropdownClicked(false)"
					@keydown.enter="dropdownClicked(true)"
				>
					<div class="truncate">{{ selectedItemNames ?? 'Auswählen...' }}</div>
					<ArrowDown class="shrink-0 grow-0 duration-100 w-4" :class="{'rotate-180': props.ui.expanded.v}" />
				</div>
			</InputWrapper>

			<!-- Dropdown content -->
			<div ref="dropdownOptions">
				<div v-if="props.ui.expanded.v" class="absolute top-full left-0 right-0 bg-white shadow-lg mt-1 z-40">
					<ui-dropdown-item
						v-for="(dropdownItem, index) in props.ui.items.v"
						:key="index"
						:ui="dropdownItem"
						:multiselect="props.ui.multiselect.v"
						:selected="isSelected(dropdownItem)"
					/>
					<div v-if="props.ui.multiselect.v" class="flex justify-center p-2">
						<button class="button-primary w-full max-w-64" @click="dropdownClicked(true)">Schließen</button>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>
