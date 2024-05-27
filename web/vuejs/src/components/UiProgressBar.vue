<template>
	<progress v-if="ui.max.v && ui.value.v" :max="ui.max.v" :value="ui.value.v"></progress>
	<!-- TODO: Style indeterminate progress bar -->
	<progress v-else></progress>
</template>

<script setup lang="ts">
import type { ProgressBar } from '@/shared/protocol/ora/progressBar';

defineProps<{
	ui: ProgressBar;
}>();
</script>
