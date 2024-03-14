<script setup lang="ts">
import type { LiveDropdown } from '@/shared/model';
import UiDropdownItem from '@/components/UiDropdownItem.vue';
import { computed } from 'vue';

const props = defineProps<{
  ui: LiveDropdown;
}>();

const selectedItemName = computed((): string => {
	return props.ui.items.value.find((item) => item.itemIndex.value === props.ui.selectedIndex.value)?.content.value ?? '';
});
</script>

<template>
	<div>
		<p>Selected: {{ selectedItemName }}</p>
		<div class="flex flex-col">
			<ui-dropdown-item v-for="(dropdownItem, index) in props.ui.items.value" :key="index" :ui="dropdownItem" />
		</div>
	</div>
</template>
