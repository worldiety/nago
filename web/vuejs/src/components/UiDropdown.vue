<script setup lang="ts">
import type { LiveDropdown } from '@/shared/model';
import UiDropdownItem from '@/components/UiDropdownItem.vue';

const props = defineProps<{
  ui: LiveDropdown;
}>();
</script>

<template>
	<div class="flex flex-col">
		<ui-dropdown-item v-for="(dropdownItem, index) in props.ui.items.value" :key="index" :ui="dropdownItem" />
	</div>
</template>
