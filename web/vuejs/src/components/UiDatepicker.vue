<script setup lang="ts">
import type { LiveDatepicker } from '@/shared/model/liveDatepicker';

const props = defineProps<{
	ui: LiveDatepicker;
}>();
</script>

<template>
	<p>{{ props.ui.type }}</p>
</template>
