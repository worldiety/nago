<template>
	<div class="relative" >
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
				<div v-if="selectedItemNames" class="truncate text-black pr-6">{{ selectedItemNames}}</div>
				<div v-else class="truncate text-placeholder-text">{{ $t('dropdown.select') }}</div>
				<ArrowDown class="absolute shrink-0 grow-0 duration-100 w-3.5 right-3" :class="{'rotate-180': props.ui.expanded.v}" />
			</div>
		</InputWrapper>

		<!-- Dropdown content -->
		<div ref="dropdownOptions">
			<div v-if="props.ui.expanded.v" class="absolute bg-background top-full left-0 right-0 shadow-ora-shadow rounded-2lg mt-2.5 py-2.5 z-40">
				<ui-dropdown-searchfilter  v-if="props.ui.searchable.v" @searchQueryChanged="(updatedSearchQuery) => searchQuery = updatedSearchQuery"></ui-dropdown-searchfilter>
				<ui-dropdown-item
					v-for="(dropdownItem, index) in itemsFiltered"
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
</template>

<script setup lang="ts">
import UiDropdownItem from '@/components/dropdown/UiDropdownItem.vue';
import ArrowDown from '@/assets/svg/arrowDown.svg';
import { computed, onMounted, onUpdated, ref } from 'vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import type {Dropdown} from "@/shared/protocol/ora/dropdown";
import type {DropdownItem} from "@/shared/protocol/ora/dropdownItem";
import UiDropdownSearchfilter from "@/components/dropdown/UiDropdownSearchfilter.vue";
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: Dropdown;
}>();

const serviceAdapter = useServiceAdapter();
const dropdownOptions = ref<HTMLElement|undefined>();
const searchQuery = ref<string>("");

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

const itemsFiltered = computed((): DropdownItem[] => {
	const searchTerms = searchQuery.value.toLowerCase().trim().split(/\s+/);
	return props.ui.items.v?.filter((item: DropdownItem) => {
		const combinedItem = item.content.v.toLowerCase().replace(/\s+/g, "");

		return searchTerms.every(searchTerm => combinedItem.includes(searchTerm));
	});
});


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
			serviceAdapter.executeFunctions(props.ui.onClicked);
		}
	}
}

function dropdownClicked(forceClose: boolean): void {
	if (!props.ui.disabled.v && (forceClose || !props.ui.expanded.v)) {
		serviceAdapter.executeFunctions(props.ui.onClicked);
	}
	searchQuery.value = ''
}

function isSelected(item: DropdownItem): boolean {
	const itemIndex = indexOf(item);
	return props.ui.selectedIndices.v?.includes(itemIndex) ?? false;
}
</script>
