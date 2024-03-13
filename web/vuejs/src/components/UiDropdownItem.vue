<script setup lang="ts">
import type { LiveDropdownItem } from '@/shared/model';

const props = defineProps<{
	ui: LiveDropdownItem;
}>();
</script>

<template>
	<div>
		<p>{{ props.ui.content.name }}</p>
	</div>
</template>
